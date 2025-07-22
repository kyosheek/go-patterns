package factory

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

// positive tests

// sample is a non-comparable struct used to make sure the generic
// implementation also works for user defined types.
type sample struct {
	A int
	B string
}

// zero returns the zero value for any type T.  Useful to avoid having
// to repeat “0”, “” or “sample{}” everywhere in the tests.
func zero[T any]() (z T) { return }

func TestCreate(t *testing.T) {
	t.Run("int – zero value", func(t *testing.T) {
		fac := New[int]()
		got := fac.Create()
		if got != 0 {
			t.Fatalf("Create() = %v, want 0", got)
		}
	},
	)

	t.Run("string – zero value", func(t *testing.T) {
		fac := New[string]()
		got := fac.Create()
		if got != "" {
			t.Fatalf(`Create() = %q, want ""`, got)
		}
	},
	)

	t.Run("struct – zero value", func(t *testing.T) {
		fac := New[sample]()
		got := fac.Create()
		if !reflect.DeepEqual(got, zero[sample]()) {
			t.Fatalf("Create() = %#v, want %#v", got, zero[sample]())
		}
	},
	)
}

func TestGetShared(t *testing.T) {
	t.Run("same int factory returns identical pointer", func(t *testing.T) {
		fac := New[int]()

		p1 := fac.GetShared()
		p2 := fac.GetShared()

		if p1 != p2 {
			t.Fatalf("GetShared() returned different addresses: %p vs %p", p1, p2)
		}

		// Mutating through one pointer must be visible through the other.
		*p1 = 42
		if *p2 != 42 {
			t.Fatalf("mutation through shared pointer not visible, got %d, want 42", *p2)
		}
	},
	)

	t.Run("same sample factory returns identical pointer", func(t *testing.T) {
		fac := New[sample]()

		p1 := fac.GetShared()
		p2 := fac.GetShared()

		if p1 != p2 {
			t.Fatalf("GetShared() returned different addresses: %p vs %p", p1, p2)
		}

		// Mutating through one pointer must be visible through the other.
		p1.A = 42
		p1.B = "is life"
		if p2.A != 42 || p2.B != "is life" {
			t.Fatalf("mutation through shared pointer not visible, got %d, want 42, or got %s, want is life",
				p2.A,
				p2.B,
			)
		}
	},
	)

	t.Run("different int factories hold independent singletons", func(t *testing.T) {
		f1 := New[int]()
		f2 := New[int]()

		p1 := f1.GetShared()
		p2 := f2.GetShared()

		if p1 == p2 {
			t.Fatalf("distinct int factories unexpectedly share the same singleton pointer: %p", p1)
		}

		*p1 = 13
		*p2 = 37

		if *p1 == *p2 {
			t.Fatalf("values leaked between int factories: %d == %d", *p1, *p2)
		}
	},
	)

	t.Run("different sample factories hold independent singletons", func(t *testing.T) {
		f1 := New[sample]()
		f2 := New[sample]()

		p1 := f1.GetShared()
		p2 := f2.GetShared()

		if p1 == p2 {
			t.Fatalf("distinct sample factories unexpectedly share the same singleton pointer: %p", p1)
		}

		p1.A = 13
		p1.B = "is bad"
		p2.A = 37
		p2.B = "is leet"

		if p1.A == p2.A || p1.B == p2.B {
			t.Fatalf("values leaked between sample factories: %d == %d or %s == %s", p1.A, p2.A, p1.B, p2.B)
		}
	},
	)
}

func TestGetSharedIntThreadSafety(t *testing.T) {
	const goroutines = 100

	fac := New[int]()

	// We let the first goroutine set the value; everyone else must see it.
	var once sync.Once
	var expected int32 = 99

	// We also count how many distinct *int pointers we encounter.
	var uniquePtr atomic.Pointer[int]
	var differentAddress int32 // increments if any goroutine observes a different address

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			p := fac.GetShared()

			// Ensure all goroutines see the exact same pointer.
			if uniquePtr.Load() == nil {
				uniquePtr.Store(p)
			} else if uniquePtr.Load() != p {
				atomic.AddInt32(&differentAddress, 1)
			}

			// The first goroutine sets the value.
			once.Do(func() { *p = int(expected) })

			// Everyone should read the same value.
			if v := *p; int32(v) != expected {
				t.Errorf("observed value %d, want %d", v, expected)
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt32(&differentAddress) != 0 {
		t.Fatalf("GetShared() returned different pointers across goroutines")
	}
}

func TestGetSharedSampleThreadSafety(t *testing.T) {
	const goroutines = 100

	fac := New[sample]()

	// We let the first goroutine set the value; everyone else must see it.
	var once sync.Once
	var expected sample = sample{
		A: 42,
		B: "is life",
	}

	// We also count how many distinct *int pointers we encounter.
	var uniquePtr atomic.Pointer[sample]
	var differentAddress int32 // increments if any goroutine observes a different address

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()

			p := fac.GetShared()

			// Ensure all goroutines see the exact same pointer.
			if uniquePtr.Load() == nil {
				uniquePtr.Store(p)
			} else if uniquePtr.Load() != p {
				atomic.AddInt32(&differentAddress, 1)
			}

			// The first goroutine sets the value.
			once.Do(func() { *p = expected })

			// Everyone should read the same value.
			if v := *p; v != expected {
				t.Errorf("observed value %v, want %v", v, expected)
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt32(&differentAddress) != 0 {
		t.Fatalf("GetShared() returned different pointers across goroutines")
	}
}

func TestNew(t *testing.T) {
	t.Run("returned object implements Factory", func(t *testing.T) {
		fac := New[string]()

		// Compile-time assertion – the following assignment will fail to
		// compile if fac does not implement Factory[string].
		var _ Factory[string] = fac
	},
	)

	t.Run("creates usable factory", func(t *testing.T) {
		fac := New[string]()
		if got := fac.Create(); got != "" {
			t.Fatalf("Create() on fresh factory = %q, want empty string", got)
		}
		if ptr := fac.GetShared(); ptr == nil {
			t.Fatalf("GetShared() returned nil pointer")
		}
	},
	)
}

// negative tests

// ptrEqual works around the fact that we cannot compare pointers of
// different concrete types directly.  We go through unsafe.Pointer which
// is comparable.
func ptrEqual(a, b any) bool {
	return unsafe.Pointer(reflect.ValueOf(a).Pointer()) ==
		unsafe.Pointer(reflect.ValueOf(b).Pointer())
}

func TestFactoriesWithDifferentTypesAreIndependent(t *testing.T) {
	fInt := New[int]()
	fStr := New[string]()
	fSlice := New[[]byte]()

	pInt := fInt.GetShared()
	pStr := fStr.GetShared()
	pSlc := fSlice.GetShared()

	if ptrEqual(pInt, pStr) || ptrEqual(pInt, pSlc) || ptrEqual(pStr, pSlc) {
		t.Fatalf("singleton instances leaked across generic types – addresses collide")
	}
}

func TestPointerElementTypeBehaviour(t *testing.T) {
	type pointer = *int

	f := New[pointer]()

	shared := f.GetShared()
	if shared == nil {
		t.Fatalf("GetShared() returned <nil> pointer (expected non-nil pointer TO a nil *int)")
	}
	if *shared != nil {
		t.Fatalf("initial **int value is not nil; got %v", *shared)
	}

	val := new(int)
	*val = 73
	*shared = val

	if got := *f.GetShared(); got != val {
		t.Fatalf("mutation not persisted through singleton: got %v want %v", got, val)
	}
}

func TestCreateReturnsProperZeroForOddTypes(t *testing.T) {
	type oddTestCase[T any] struct {
		name     string
		expected T
	}

	tests := []any{
		oddTestCase[[]int]{name: "slice", expected: nil},
		oddTestCase[map[string]int]{name: "map", expected: nil},
		oddTestCase[func()]{name: "func", expected: nil},
	}

	for _, raw := range tests {
		switch tc := raw.(type) {
		case oddTestCase[[]int]:
			f := New[[]int]()
			if got := f.Create(); !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("%s zero value mismatch: got %#v, want %#v", tc.name, got, tc.expected)
			}
		case oddTestCase[map[string]int]:
			f := New[map[string]int]()
			if got := f.Create(); !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("%s zero value mismatch: got %#v, want %#v", tc.name, got, tc.expected)
			}
		case oddTestCase[func()]:
			f := New[func()]()
			if got := f.Create(); !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("%s zero value mismatch: got %p, want %p", tc.name, got, tc.expected)
			}
		}
	}
}

func TestCreateDoesNotAffectSingleton(t *testing.T) {
	f := New[int]()

	shared := f.GetShared()
	*shared = 21

	if got := f.Create(); got != 0 {
		t.Fatalf("Create() leaked singleton value: got %d, want 0", got)
	}
	if *shared != 21 {
		t.Fatalf("singleton value was mutated by Create(): got %d, want 21", *shared)
	}
}

func TestConcurrentAccessDifferentTypes(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	fInt := New[int]()
	fStr := New[string]()

	go func() {
		defer wg.Done()
		p := fInt.GetShared()
		*p = 1984
		if v := *fInt.GetShared(); v != 1984 {
			t.Errorf("int factory lost its value under concurrency: got %d", v)
		}
	}()

	go func() {
		defer wg.Done()
		p := fStr.GetShared()
		*p = "cypher"
		if v := *fStr.GetShared(); v != "cypher" {
			t.Errorf("string factory lost its value under concurrency: got %q", v)
		}
	}()

	wg.Wait()
}
