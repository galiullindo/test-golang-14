package internal

import "sync"

type Statistics struct {
	mu     sync.Mutex
	ok     int
	errors int
}

func (s *Statistics) OK() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ok
}
func (s *Statistics) Errors() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.errors
}
func (s *Statistics) IncrementOK() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ok++
}
func (s *Statistics) IncrementErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors++
}
