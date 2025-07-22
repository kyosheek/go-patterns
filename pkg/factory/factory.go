package factory

import (
	"kyoshee/patterns/pkg/singleton"
)

type Factory[T any] interface {
	Create() T
	GetShared() *T
}

type ConcreteFactory[T any] struct {
	shared *singleton.Singleton[T]
}

func New[T any]() Factory[T] {
	return &ConcreteFactory[T]{
		shared: singleton.New(func() *T {
			var t T
			return &t
		},
		),
	}
}

func (f *ConcreteFactory[T]) Create() T {
	var t T
	return t
}

func (f *ConcreteFactory[T]) GetShared() *T {
	return f.shared.Get()
}
