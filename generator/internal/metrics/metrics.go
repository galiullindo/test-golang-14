package metrics

import "sync"

// RequestStatistics ...
type RequestStatistics struct {
	mu     sync.Mutex
	ok     int
	errors int
}

// NewRequestStatistics ...
func NewRequestStatistics() *RequestStatistics {
	return &RequestStatistics{}
}

// OK ...
func (s *RequestStatistics) OK() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ok
}

// Errors ...
func (s *RequestStatistics) Errors() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.errors
}

// IncrementOK ...
func (s *RequestStatistics) IncrementOK() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ok++
}

// IncrementErrors ...
func (s *RequestStatistics) IncrementErrors() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors++
}
