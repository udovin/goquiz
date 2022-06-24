// Package models contains tools for working with GoQuiz objects stored
// in different databases like SQLite or Postgres.
package models

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/udovin/gosql"
	"github.com/udovin/solve/db"
	"github.com/udovin/solve/models"
)

// Cloner represents object that can be cloned.
type Cloner[T any] interface {
	Clone() T
}

type (
	NInt64  = models.NInt64
	JSON    = models.JSON
	NString = models.NString
)

type index[K comparable] map[K]map[int64]struct{}

func makeIndex[K comparable]() index[K] {
	return map[K]map[int64]struct{}{}
}

func (m index[K]) Create(key K, id int64) {
	if _, ok := m[key]; !ok {
		m[key] = map[int64]struct{}{}
	}
	m[key][id] = struct{}{}
}

func (m index[K]) Delete(key K, id int64) {
	delete(m[key], id)
	if len(m[key]) == 0 {
		delete(m, key)
	}
}

type pair[F, S any] struct {
	First  F
	Second S
}

func makePair[F, S any](f F, s S) pair[F, S] {
	return pair[F, S]{First: f, Second: s}
}

// EventType represents type of object event.
type EventType int8

const (
	// CreateEvent means that this is event of object creation.
	CreateEvent EventType = 1
	// DeleteEvent means that this is event of object deletion.
	DeleteEvent EventType = 2
	// UpdateEvent means that this is event of object modification.
	UpdateEvent EventType = 3
)

// String returns string representation of event.
func (t EventType) String() string {
	switch t {
	case CreateEvent:
		return "create"
	case DeleteEvent:
		return "delete"
	case UpdateEvent:
		return "update"
	default:
		return fmt.Sprintf("EventType(%d)", t)
	}
}

// ObjectEvent represents event for object.
type ObjectEvent[T db.Object] interface {
	db.Event
	// EventType should return type of object event.
	EventType() EventType
	// Object should return struct with object data.
	Object() T
}

type ObjectEventPtr[T db.Object] interface {
	ObjectEvent[T]
	// SetObject should replace event object.
	SetObject(T)
	// SetAccountID should set account ID for specified object event.
	SetAccountID(int64)
}

// baseEvent represents base for all events.
type baseEvent struct {
	// BaseEventID contains event id.
	BaseEventID int64 `db:"event_id"`
	// BaseEventType contains type of event.
	BaseEventType EventType `db:"event_type"`
	// BaseEventTime contains event type.
	BaseEventTime int64 `db:"event_time"`
	// EventAccountID contains account id.
	EventAccountID NInt64 `db:"event_account_id"`
}

// EventId returns id of this event.
func (e baseEvent) EventID() int64 {
	return e.BaseEventID
}

// EventTime returns time of this event.
func (e baseEvent) EventTime() time.Time {
	return time.Unix(e.BaseEventTime, 0)
}

// EventType returns type of this event.
func (e baseEvent) EventType() EventType {
	return e.BaseEventType
}

func (e *baseEvent) SetAccountID(accountID int64) {
	e.EventAccountID = NInt64(accountID)
}

type accountIDKey struct{}

func WithAccountID(ctx context.Context, id int64) context.Context {
	return context.WithValue(ctx, accountIDKey{}, id)
}

// GetAccountID returns account ID or zero if there is no account.
func GetAccountID(ctx context.Context) int64 {
	if id, ok := ctx.Value(accountIDKey{}).(int64); ok {
		return id
	}
	return 0
}

// makeBaseEvent creates baseEvent with specified type.
func makeBaseEvent(t EventType) baseEvent {
	return baseEvent{BaseEventType: t, BaseEventTime: time.Now().Unix()}
}

type baseStoreImpl[T db.Object, E ObjectEvent[T]] interface {
	reset()
	makeObject(id int64) T
	makeObjectEvent(EventType) E
	onCreateObject(T)
	onDeleteObject(int64)
	onUpdateObject(T)
}

// Store represents cached store.
type Store interface {
	Init(ctx context.Context) error
	Sync(ctx context.Context) error
}

type baseStore[T db.Object, E ObjectEvent[T]] struct {
	db       *gosql.DB
	table    string
	objects  db.ObjectStore[T]
	events   db.EventStore[E]
	consumer db.EventConsumer[E]
	impl     baseStoreImpl[T, E]
	mutex    sync.RWMutex
}

// DB returns store database.
func (s *baseStore[T, E]) DB() *gosql.DB {
	return s.db
}

func (s *baseStore[T, E]) Init(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.initUnlocked(ctx)
}

func (s *baseStore[T, E]) initUnlocked(ctx context.Context) error {
	if tx := db.GetTx(ctx); tx == nil {
		return gosql.WrapTx(ctx, s.db, func(tx *sql.Tx) error {
			return s.initUnlocked(db.WithTx(ctx, tx))
		}, sqlReadOnly)
	}
	if err := s.initEvents(ctx); err != nil {
		return err
	}
	return s.initObjects(ctx)
}

const eventGapSkipWindow = 25000

func (s *baseStore[T, E]) initEvents(ctx context.Context) error {
	beginID, err := s.events.LastEventID(ctx)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
		beginID = 1
	}
	if beginID > eventGapSkipWindow {
		beginID -= eventGapSkipWindow
	} else {
		beginID = 1
	}
	s.consumer = db.NewEventConsumer[E](s.events, beginID)
	return s.consumer.ConsumeEvents(ctx, func(E) error {
		return nil
	})
}

func (s *baseStore[T, E]) initObjects(ctx context.Context) error {
	rows, err := s.objects.LoadObjects(ctx)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	s.impl.reset()
	for rows.Next() {
		s.impl.onCreateObject(rows.Row())
	}
	return rows.Err()
}

func (s *baseStore[T, E]) Sync(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.consumer.ConsumeEvents(ctx, s.consumeEvent)
}

// Create creates object and returns copy with valid ID.
func (s *baseStore[T, E]) Create(ctx context.Context, object *T) error {
	event := s.impl.makeObjectEvent(CreateEvent)
	eventPtr := any(&event).(ObjectEventPtr[T])
	eventPtr.SetObject(*object)
	if err := s.createObjectEvent(ctx, &event); err != nil {
		return err
	}
	*object = event.Object()
	return nil
}

// Update updates object with specified ID.
func (s *baseStore[T, E]) Update(ctx context.Context, object T) error {
	event := s.impl.makeObjectEvent(UpdateEvent)
	eventPtr := any(&event).(ObjectEventPtr[T])
	eventPtr.SetObject(object)
	return s.createObjectEvent(ctx, &event)
}

// Delete deletes compiler with specified ID.
func (s *baseStore[T, E]) Delete(ctx context.Context, id int64) error {
	event := s.impl.makeObjectEvent(DeleteEvent)
	eventPtr := any(&event).(ObjectEventPtr[T])
	eventPtr.SetObject(s.impl.makeObject(id))
	return s.createObjectEvent(ctx, &event)
}

var (
	sqlRepeatableRead = gosql.WithIsolation(sql.LevelRepeatableRead)
	sqlReadOnly       = gosql.WithReadOnly(true)
)

func (s *baseStore[T, E]) createObjectEvent(
	ctx context.Context, event *E,
) error {
	// Force creation of new transaction.
	if tx := db.GetTx(ctx); tx == nil {
		return gosql.WrapTx(ctx, s.db, func(tx *sql.Tx) error {
			return s.createObjectEvent(db.WithTx(ctx, tx), event)
		}, sqlRepeatableRead)
	}
	eventPtr := any(event).(ObjectEventPtr[T])
	eventPtr.SetAccountID(GetAccountID(ctx))
	switch object := eventPtr.Object(); eventPtr.EventType() {
	case CreateEvent:
		if err := s.objects.CreateObject(ctx, &object); err != nil {
			return err
		}
		eventPtr.SetObject(object)
	case UpdateEvent:
		if err := s.objects.UpdateObject(ctx, &object); err != nil {
			return err
		}
		eventPtr.SetObject(object)
	case DeleteEvent:
		if err := s.objects.DeleteObject(ctx, object.ObjectID()); err != nil {
			return err
		}
	}
	return s.events.CreateEvent(ctx, event)
}

func (s *baseStore[T, E]) lockStore(tx *sql.Tx) error {
	switch s.db.Dialect() {
	case gosql.SQLiteDialect:
		return nil
	default:
		_, err := tx.Exec(fmt.Sprintf("LOCK TABLE %q", s.table))
		return err
	}
}

func (s *baseStore[T, E]) onUpdateObject(object T) {
	s.impl.onDeleteObject(object.ObjectID())
	s.impl.onCreateObject(object)
}

func (s *baseStore[T, E]) consumeEvent(e E) error {
	switch object := e.Object(); e.EventType() {
	case CreateEvent:
		s.impl.onCreateObject(object)
	case DeleteEvent:
		s.impl.onDeleteObject(object.ObjectID())
	case UpdateEvent:
		s.impl.onUpdateObject(object)
	default:
		return fmt.Errorf("unexpected event type: %v", e.EventType())
	}
	return nil
}

func makeBaseStore[T db.Object, E ObjectEvent[T]](
	conn *gosql.DB,
	table, eventTable string,
	impl baseStoreImpl[T, E],
) baseStore[T, E] {
	var event E
	if _, ok := any(&event).(ObjectEventPtr[T]); !ok {
		panic(fmt.Errorf("event %T does not implement ObjectEventPtr[T]", event))
	}
	return baseStore[T, E]{
		db:      conn,
		table:   table,
		objects: db.NewObjectStore[T]("id", table, conn),
		events:  db.NewEventStore[E]("event_id", eventTable, conn),
		impl:    impl,
	}
}
