package observer

import (
	"testing"
)

func TestSubject_Attach_NoObservers(t *testing.T) {
	t.Parallel()

	s := NewSubject[int]()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Attach with no observers panicked: %v", r)
		}
	}()
	s.Attach()

	if len(s.observers) != 0 {
		t.Fatalf("expected 0 observers, got %d", len(s.observers))
	}
}

func TestSubject_SetState_NoObservers(t *testing.T) {
	t.Parallel()

	s := NewSubject[int]()
	const state = 99
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SetState with no observers panicked: %v", r)
		}
	}()
	s.SetState(state)

	if s.state != state {
		t.Fatalf("state = %d, want %d", s.state, state)
	}
}

func TestSubject_SetState_WithNilObserver_ShouldPanic(t *testing.T) {
	t.Parallel()

	s := NewSubject[int]()
	var nilObs Observer[int]
	s.Attach(nilObs)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when notifying a nil observer, got none")
		}
	}()
	s.SetState(1)
}

func TestNilSubjectMethods_ShouldPanic(t *testing.T) {
	t.Parallel()

	var s *Subject[int] // nil subject

	check := func(name string, f func()) {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("%s did not panic on nil receiver", name)
			}
		}()
		f()
	}

	check("Attach", func() { s.Attach() })
	check("SetState", func() { s.SetState(1) })
	check("notifyAll", func() { s.notifyAll() })
}
