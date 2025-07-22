package factory

import (
	"kyoshee/patterns/pkg/singleton"
)

type GenericFactory[T any] interface {
	Create() T
	GetShared() *T
}

type Factory[T any] struct {
	shared *singleton.Singleton[T]
}

func New[T any]() GenericFactory[T] {
	return &Factory[T]{
		shared: singleton.New(func() *T {
			var t T
			return &t
		},
		),
	}
}

func (f *Factory[T]) Create() T {
	var t T
	return t
}

func (f *Factory[T]) GetShared() *T {
	return f.shared.Get()
}
