package factory

import (
	"reflect"
	"testing"
)

// sample is a non-comparable struct used to make sure the generic
// implementation also works for user defined types.
type sample struct {
	A int
	B string
}

// zero returns the zero value for any type T.  Useful to avoid having
// to repeat “0”, “” or “sample{}” everywhere in the tests.
func zero[T any]() (z T) { return }

func TestFactory_Create(t *testing.T) {
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

func TestFactory_GetShared(t *testing.T) {
	t.Run("same factory returns identical pointer", func(t *testing.T) {
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

	t.Run("different factories hold independent singletons", func(t *testing.T) {
		f1 := New[int]()
		f2 := New[int]()

		p1 := f1.GetShared()
		p2 := f2.GetShared()

		if p1 == p2 {
			t.Fatalf("distinct factories unexpectedly share the same singleton pointer: %p", p1)
		}

		*p1 = 13
		*p2 = 37

		if *p1 == *p2 {
			t.Fatalf("values leaked between factories: %d == %d", *p1, *p2)
		}
	},
	)
}

func TestNew(t *testing.T) {
	t.Run("returned object implements GenericFactory", func(t *testing.T) {
		fac := New[string]()

		// Compile-time assertion – the following assignment will fail to
		// compile if fac does not implement GenericFactory[string].
		var _ GenericFactory[string] = fac
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
