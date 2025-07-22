package observer

type Observer[T any] interface {
	Update(state T)
}

type Subject[T any] struct {
	observers []Observer[T]
	state     T
}

func (s *Subject[T]) Attach(observers ...Observer[T]) {
	if s == nil {
		panic("subject is not initialized")
	}

	for _, observer := range observers {
		s.observers = append(s.observers, observer)
	}
}

func (s *Subject[T]) SetState(state T) {
	if s == nil {
		panic("subject is not initialized")
	}

	s.state = state
	s.notifyAll()
}

func (s *Subject[T]) notifyAll() {
	for _, observer := range s.observers {
		observer.Update(s.state)
	}
}

func NewSubject[T any]() *Subject[T] {
	var s T
	return &Subject[T]{
		observers: make([]Observer[T], 0),
		state:     s,
	}
}
