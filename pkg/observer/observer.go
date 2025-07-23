package observer

// Observer interface requires Update function that will run
// on concrete instance when Subject receives new state
type Observer[T any] interface {
	Update(state T, prevState T)
}

// Subject is a struct that holds multiple Observer instances
// that need to react on change of given struct state
type Subject[T any] struct {
	observers []Observer[T]
	state     T
}

// Attach adds new Observer instances to current Subject.
// Subject must be initialized before Attach calls.
func (s *Subject[T]) Attach(observers ...Observer[T]) {
	if s == nil {
		panic("subject is not initialized")
	}

	for _, observer := range observers {
		s.observers = append(s.observers, observer)
	}
}

// SetState updates current Subject state and notifies all attached Observer instances.
// Subject must be initialized before Attach calls.
func (s *Subject[T]) SetState(state T) {
	if s == nil {
		panic("subject is not initialized")
	}

	prevState := s.state
	s.state = state
	s.notifyAll(prevState)
}

// notifyAll is a helper function that calls Update on attached Observer instances.
// Since notifyAll called only inside SetState calls, this function does not
// directly check for instantiation of current Subject.
func (s *Subject[T]) notifyAll(prevState T) {
	for _, observer := range s.observers {
		observer.Update(s.state, prevState)
	}
}

// NewSubject creates new Subject with given state type.
// Subject can attach Observer instances
// and set new State.
func NewSubject[T any]() *Subject[T] {
	var s T
	return &Subject[T]{
		observers: make([]Observer[T], 0),
		state:     s,
	}
}
