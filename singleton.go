package singleton

import (
	"sync"
)

type Singleton[T any] struct {
	once  sync.Once
	f     func() *T
	value *T
}

func New[T any](f func() *T) *Singleton[T] {
	return &Singleton[T]{
		once:  sync.Once{},
		f:     f,
		value: nil,
	}
}

func (s *Singleton[T]) Get() *T {
	s.once.Do(func() {
		s.value = s.f()
		s.f = nil
	},
	)
	return s.value
}
