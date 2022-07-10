package models

import (
	"github.com/udovin/gosql"
)

type Quiz struct {
	baseObject
}

type QuizEvent struct {
	baseEvent
	Quiz
}

func (e QuizEvent) Object() Quiz {
	return e.Quiz
}

func (e *QuizEvent) SetObject(o Quiz) {
	e.Quiz = o
}

type QuizStore struct {
	baseStore[Quiz, QuizEvent, *Quiz, *QuizEvent]
	quizes map[int64]Quiz
}

func (s *QuizStore) reset() {
	s.quizes = map[int64]Quiz{}
}

func (s *QuizStore) onCreateObject(quiz Quiz) {
	s.quizes[quiz.ID] = quiz
}

func (s *QuizStore) onDeleteObject(id int64) {
	if quiz, ok := s.quizes[id]; ok {
		delete(s.quizes, quiz.ID)
	}
}

// NewQuizStore creates a new instance of QuizStore.
func NewQuizStore(
	db *gosql.DB, table, eventTable string,
) *QuizStore {
	impl := &QuizStore{}
	impl.baseStore = makeBaseStore[Quiz, QuizEvent](
		db, table, eventTable, impl,
	)
	return impl
}
