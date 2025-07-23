package observer

import (
	"reflect"
	"sync"
	"testing"
)

// positive tests

// testObserver is a helper Observer implementation that stores every state it receives.
// The internal slice is guarded by a mutex so we can safely use it from
// several goroutines when we run the thread-safety test.
type testObserver struct {
	mu     sync.Mutex
	states []int
}

func (o *testObserver) Update(state, _ int) {
	o.mu.Lock()
	o.states = append(o.states, state)
	o.mu.Unlock()
}

func (o *testObserver) lastState(t *testing.T) (int, bool) {
	t.Helper()

	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.states) == 0 {
		return 0, false
	}
	return o.states[len(o.states)-1], true
}

func TestNew(t *testing.T) {
	t.Parallel()

	got := NewSubject[int]()

	want := &Subject[int]{
		observers: make([]Observer[int], 0),
		state:     0,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NewSubject[int]() = %+v, want %+v", got, want)
	}
}

func TestAttach(t *testing.T) {
	t.Parallel()

	s := NewSubject[int]()
	o1 := &testObserver{}
	o2 := &testObserver{}

	s.Attach(o1, o2)

	expected := []Observer[int]{o1, o2}
	if !reflect.DeepEqual(s.observers, expected) {
		t.Fatalf("Attach() resulted in observers %#v, want %#v", s.observers, expected)
	}
}

func TestSetState(t *testing.T) {
	t.Parallel()

	s := NewSubject[int]()
	o := &testObserver{}
	s.Attach(o)

	state := 42
	s.SetState(state)

	// Subject’s own state should have been updated.
	if s.state != state {
		t.Fatalf("state after SetState(%d) = %d, want %d", state, s.state, state)
	}

	// Observer should have been notified exactly once with the same value.
	if got, ok := o.lastState(t); !ok || got != state {
		t.Fatalf("observer received %d, want %d", got, state)
	}
	if len(o.states) != 1 {
		t.Fatalf("observer should have 1 update, got %d", len(o.states))
	}
}

func TestNotifyAll(t *testing.T) {
	t.Parallel()

	s := NewSubject[int]()
	o1 := &testObserver{}
	o2 := &testObserver{}
	s.Attach(o1, o2)

	const state = 7
	s.state = state
	s.notifyAll(state)

	for i, o := range []*testObserver{o1, o2} {
		if got, ok := o.lastState(t); !ok || got != state {
			t.Fatalf("observer %d received %d, want %d", i+1, got, state)
		}
	}
}

func TestThreadSafety(t *testing.T) {
	t.Parallel()

	const (
		subjects      = 32
		updatesPerSub = 64
	)

	var wg sync.WaitGroup
	for i := 0; i < subjects; i++ {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			s := NewSubject[int]()
			o := &testObserver{}
			s.Attach(o)

			for j := 0; j < updatesPerSub; j++ {
				s.SetState(j)
			}

			if len(o.states) != updatesPerSub {
				t.Errorf("subject %d: observer received %d updates, want %d",
					i, len(o.states), updatesPerSub,
				)
				return
			}
		}(i)
	}
	wg.Wait()
}

// negative tests

func TestAttachNoObservers(t *testing.T) {
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

func TestSetStateNoObservers(t *testing.T) {
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

func TestSetStateWithNilObserverShouldPanic(t *testing.T) {
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

func TestNilSubjectMethodsShouldPanic(t *testing.T) {
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
	check("notifyAll", func() { s.notifyAll(0) })
}
