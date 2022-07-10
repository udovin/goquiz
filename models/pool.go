package models

import (
	"github.com/udovin/gosql"
)

type Pool struct {
	baseObject
}

type PoolEvent struct {
	baseEvent
	Pool
}

func (e PoolEvent) Object() Pool {
	return e.Pool
}

func (e *PoolEvent) SetObject(o Pool) {
	e.Pool = o
}

type PoolStore struct {
	baseStore[Pool, PoolEvent, *Pool, *PoolEvent]
	pools map[int64]Pool
}

func (s *PoolStore) reset() {
	s.pools = map[int64]Pool{}
}

func (s *PoolStore) onCreateObject(pool Pool) {
	s.pools[pool.ID] = pool
}

func (s *PoolStore) onDeleteObject(id int64) {
	if pool, ok := s.pools[id]; ok {
		delete(s.pools, pool.ID)
	}
}

// NewPoolStore creates a new instance of PoolStore.
func NewPoolStore(
	db *gosql.DB, table, eventTable string,
) *PoolStore {
	impl := &PoolStore{}
	impl.baseStore = makeBaseStore[Pool, PoolEvent](
		db, table, eventTable, impl,
	)
	return impl
}
