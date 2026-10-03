package jobs

import (
	"errors"
	"sync"
	"time"
)

var ErrJobNotFound = errors.New("job not found")

type Result struct {
	Job       Job
	Output    []byte
	Filename  string
	ExpiresAt time.Time
}

type Store struct {
	mu      sync.RWMutex
	results map[string]Result
}

func NewStore() *Store {
	return &Store{
		results: make(map[string]Result),
	}
}

func (s *Store) Set(result Result) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results[result.Job.ID] = result
}

func (s *Store) Get(id string) (Result, error) {
	s.mu.RLock()
	result, ok := s.results[id]
	s.mu.RUnlock()

	if !ok {
		return Result{}, ErrJobNotFound
	}

	if time.Now().UTC().After(result.ExpiresAt) {
		s.Delete(id)
		return Result{}, ErrJobNotFound
	}

	return result, nil
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.results, id)
}
