package singleton

import (
	"sync"
)

// Factory is a function that returns pointer to instance of T.
type Factory[T any] func() *T

// Singleton accepts factory f and returns result of single
// execution of given factory on Get calls.
type Singleton[T any] struct {
	mu    sync.Mutex
	f     Factory[T]
	value *T
	ready bool
}

// New creates a Singleton instance with provided Factory function f.
func New[T any](f Factory[T]) *Singleton[T] {
	return &Singleton[T]{
		f: f,
	}
}

// Get safely retrieves result of single Factory function execution
// that is shared across consecutive calls. Get calls Factory function
// if current value is nil.
func (s *Singleton[T]) Get() *T {
	s.mu.Lock()
	if s.ready {
		v := s.value
		s.mu.Unlock()
		return v
	}
	f := s.f
	s.mu.Unlock()

	if f == nil {
		panic("singleton.Get(): factory function is nil")
	}

	// Run the factory *outside* the lock, so other goroutines may call
	// Get concurrently, and we don’t hold the mutex during a potentially
	// slow operation.
	v := f() // may panic; that’s fine – state hasn’t been changed yet

	s.mu.Lock()
	if !s.ready {
		s.value = v
		s.ready = true
		s.f = nil
	}
	v = s.value // value might have been stored by another goroutine
	s.mu.Unlock()

	return v
}
