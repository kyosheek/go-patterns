package singleton

import (
	"sync"
)

type Singleton[T any] struct {
	mu    sync.Mutex
	f     func() *T
	value *T
	ready bool
}

func New[T any](f func() *T) *Singleton[T] {
	return &Singleton[T]{
		f: f,
	}
}

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
