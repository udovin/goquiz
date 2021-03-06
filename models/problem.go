package models

import (
	"github.com/udovin/gosql"
)

type Problem struct {
	baseObject
}

type ProblemEvent struct {
	baseEvent
	Problem
}

func (e ProblemEvent) Object() Problem {
	return e.Problem
}

func (e *ProblemEvent) SetObject(o Problem) {
	e.Problem = o
}

type ProblemStore struct {
	baseStore[Problem, ProblemEvent, *Problem, *ProblemEvent]
	problems map[int64]Problem
}

func (s *ProblemStore) reset() {
	s.problems = map[int64]Problem{}
}

func (s *ProblemStore) onCreateObject(problem Problem) {
	s.problems[problem.ID] = problem
}

func (s *ProblemStore) onDeleteObject(id int64) {
	if problem, ok := s.problems[id]; ok {
		delete(s.problems, problem.ID)
	}
}

// NewProblemStore creates a new instance of ProblemStore.
func NewProblemStore(
	db *gosql.DB, table, eventTable string,
) *ProblemStore {
	impl := &ProblemStore{}
	impl.baseStore = makeBaseStore[Problem, ProblemEvent](
		db, table, eventTable, impl,
	)
	return impl
}
