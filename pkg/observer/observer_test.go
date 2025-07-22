package observer

import (
	"reflect"
	"sync"
	"testing"
)

/*
A helper Observer implementation that stores every state it receives.
The internal slice is guarded by a mutex so we can safely use it from
several goroutines when we run the thread-safety test.
*/
type testObserver struct {
	mu     sync.Mutex
	states []int
}

func (o *testObserver) Update(state int) {
	o.mu.Lock()
	o.states = append(o.states, state)
	o.mu.Unlock()
}

func (o *testObserver) lastState() (int, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()

	if len(o.states) == 0 {
		return 0, false
	}
	return o.states[len(o.states)-1], true
}

func TestNewSubject(t *testing.T) {
	t.Parallel()

	got := NewSubject[int]()

	// An empty int value is 0.
	want := &Subject[int]{
		observers: make([]Observer[int], 0),
		state:     0,
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("NewSubject[int]() = %+v, want %+v", got, want)
	}
}

func TestSubject_Attach(t *testing.T) {
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

func TestSubject_SetState(t *testing.T) {
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
	if got, ok := o.lastState(); !ok || got != state {
		t.Fatalf("observer received %d, want %d", got, state)
	}
	if len(o.states) != 1 {
		t.Fatalf("observer should have 1 update, got %d", len(o.states))
	}
}

func TestSubject_notifyAll(t *testing.T) {
	t.Parallel()

	s := NewSubject[int]()
	o1 := &testObserver{}
	o2 := &testObserver{}
	s.Attach(o1, o2)

	const state = 7
	s.state = state
	s.notifyAll()

	for i, o := range []*testObserver{o1, o2} {
		if got, ok := o.lastState(); !ok || got != state {
			t.Fatalf("observer %d received %d, want %d", i+1, got, state)
		}
	}
}

func TestSubject_ThreadSafety(t *testing.T) {
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
