package singleton

import (
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

const test = "test"

// positive tests

// testStringStruct is used for testing struct{v: string}.
type testStringStruct struct {
	v string
}

// testFloatStruct is used for testing struct{v: float64}.
type testFloatStruct struct {
	v float64
}

func TestNew(t *testing.T) {
	t.Parallel()

	type args[T any] struct {
		f Factory[T]
	}
	tests := []struct {
		name string
		args interface{}
	}{
		{
			name: "string",
			args: args[string]{
				f: func() *string {
					s := test
					return &s
				},
			},
		},
		{
			name: "map[uint8]uint8",
			args: args[map[uint8]uint8]{
				f: func() *map[uint8]uint8 {
					m := map[uint8]uint8{1: 2}
					return &m
				},
			},
		},
		{
			name: "[]testStringStruct",
			args: args[[]testStringStruct]{
				f: func() *[]testStringStruct {
					s := []testStringStruct{{v: test}}
					return &s
				},
			},
		},
		{
			name: "testFloatStruct",
			args: args[testFloatStruct]{
				f: func() *testFloatStruct {
					s := testFloatStruct{v: 42.0}
					return &s
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch args := tt.args.(type) {
			case args[string]:
				got := New(args.f)
				if got == nil || got.f == nil {
					t.Errorf("New() returned nil or invalid Singleton for string")
				}
			case args[map[uint8]uint8]:
				got := New(args.f)
				if got == nil || got.f == nil {
					t.Errorf("New() returned nil or invalid Singleton for map[uint8]uint8")
				}
			case args[[]testStringStruct]:
				got := New(args.f)
				if got == nil || got.f == nil {
					t.Errorf("New() returned nil or invalid Singleton for []testStringStruct")
				}
			case args[testFloatStruct]:
				got := New(args.f)
				if got == nil || got.f == nil {
					t.Errorf("New() returned nil or invalid Singleton for testFloatStruct")
				}
			}
		},
		)
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s    interface{}
		want interface{}
	}{
		{
			name: "string",
			s: New(func() *string {
				s := test
				return &s
			},
			),
			want: func() *string {
				s := test
				return &s
			}(),
		},
		{
			name: "map[uint8]uint8",
			s: New(func() *map[uint8]uint8 {
				m := map[uint8]uint8{1: 2}
				return &m
			},
			),
			want: func() *map[uint8]uint8 {
				m := map[uint8]uint8{1: 2}
				return &m
			}(),
		},
		{
			name: "[]testStringStruct",
			s: New(func() *[]testStringStruct {
				s := []testStringStruct{{v: test}}
				return &s
			},
			),
			want: func() *[]testStringStruct {
				s := []testStringStruct{{v: test}}
				return &s
			}(),
		},
		{
			name: "testFloatStruct",
			s: New(func() *testFloatStruct {
				s := testFloatStruct{v: 42.0}
				return &s
			},
			),
			want: func() *testFloatStruct {
				s := testFloatStruct{v: 42.0}
				return &s
			}(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			switch s := tt.s.(type) {
			case *Singleton[string]:
				if got := s.Get(); !reflect.DeepEqual(got, tt.want.(*string)) {
					t.Errorf("Get() = %v, want %v", got, tt.want)
				}
			case *Singleton[map[uint8]uint8]:
				if got := s.Get(); !reflect.DeepEqual(got, tt.want.(*map[uint8]uint8)) {
					t.Errorf("Get() = %v, want %v", got, tt.want)
				}
			case *Singleton[[]testStringStruct]:
				if got := s.Get(); !reflect.DeepEqual(got, tt.want.(*[]testStringStruct)) {
					t.Errorf("Get() = %v, want %v", got, tt.want)
				}
			case *Singleton[testFloatStruct]:
				if got := s.Get(); !reflect.DeepEqual(got, tt.want.(*testFloatStruct)) {
					t.Errorf("Get() = %v, want %v", got, tt.want)
				}
			}
		},
		)
	}
}

func TestThreadSafety(t *testing.T) {
	t.Parallel()

	var counter int32
	s := New(func() *int {
		atomic.AddInt32(&counter, 1)
		v := 42
		return &v
	},
	)

	var wg sync.WaitGroup
	const goroutines = 100
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			s.Get()
		}()
	}

	// Verify if singleton was created one time
	wg.Wait()
	if counter != 1 {
		t.Errorf("Singleton initialization called %d times, want 1", counter)
	}

	// Verify all goroutines get the same value
	result := *s.Get()
	for i := 0; i < goroutines; i++ {
		go func() {
			if got := *s.Get(); got != result {
				t.Errorf("Get() returned different values across goroutines")
			}
		}()
	}
}

// negative tests

func TestGetWithNilFuncPanics(t *testing.T) {
	t.Parallel()

	s := New[int](nil)

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when calling Get on Singleton created with nil func, got none")
		}
	}()

	// This should panic because the factory is nil
	_ = s.Get()
}

func TestGetWhenFactoryReturnsNil(t *testing.T) {
	t.Parallel()

	var called int32
	s := New(func() *int {
		atomic.AddInt32(&called, 1)
		return nil
	},
	)

	if got := s.Get(); got != nil {
		t.Fatalf("expected nil value from Get, got %v", got)
	}

	// A second Get must not call the factory again.
	if got := s.Get(); got != nil {
		t.Fatalf("second Get returned non-nil: %v", got)
	}

	if called != 1 {
		t.Fatalf("factory called %d times, want 1", called)
	}
}

func TestFactoryPanics(t *testing.T) {
	t.Parallel()

	var called int32
	s := New(func() *int {
		atomic.AddInt32(&called, 1)
		panic("boom")
	},
	)

	for i := 0; i < 2; i++ { // call twice, each must panic
		func() {
			defer func() {
				if r := recover(); r == nil {
					t.Fatalf("call %d: expected panic, got none", i+1)
				}
			}()
			_ = s.Get()
		}()
	}

	if called != 2 {
		t.Fatalf("factory called %d times, want 2 (once per panic)", called)
	}
}
