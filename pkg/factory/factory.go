package factory

import (
	"github.com/kyosheek/go-patterns/pkg/singleton"
)

// Factory interface requires Create and GetShared functions.
// Create creates new instance of type T.
// GetShared uses Singleton internally. See https://github.com/kyosheek/go-patterns/pkg/singleton.
type Factory[T any] interface {
	Create() T
	GetShared() *T
}

// ConcreteFactory holds shared instance for GetShared calls.
type ConcreteFactory[T any] struct {
	shared *singleton.Singleton[T]
}

// New creates ConcreteFactory with generic type T
func New[T any]() Factory[T] {
	return &ConcreteFactory[T]{
		shared: singleton.New(func() *T {
			var t T
			return &t
		},
		),
	}
}

// Create creates new instance of type T
func (f *ConcreteFactory[T]) Create() T {
	var t T
	return t
}

// GetShared returns shared instance of type T
func (f *ConcreteFactory[T]) GetShared() *T {
	return f.shared.Get()
}
