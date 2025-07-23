package throttle

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// eventually waits (with a timeout) for condition f to become true.
// avoids endless waits in case of bugs.
func eventually(t *testing.T, d time.Duration, f func() bool) {
	t.Helper()

	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if f() {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("condition not satisfied within %v", d)
}

func TestNew(t *testing.T) {
	t.Parallel()

	var calls int32
	fn := func(_ ...any) { atomic.AddInt32(&calls, 1) }

	delay := 50 * time.Millisecond
	throttled := New(fn, delay)

	// first call must go through immediately
	throttled()
	eventually(t, 20*time.Millisecond, func() bool { return atomic.LoadInt32(&calls) == 1 })

	// further calls inside the delay window must be ignored
	for i := 0; i < 10; i++ {
		throttled()
	}
	time.Sleep(20 * time.Millisecond) // still inside the 50 ms window
	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("expected 1 call inside delay window, got %d", got)
	}

	// after delay window expires the next call must go through
	time.Sleep(delay)
	throttled()
	eventually(t, 20*time.Millisecond, func() bool { return atomic.LoadInt32(&calls) == 2 })
}

func TestNewThreadSafety(t *testing.T) {
	t.Parallel()

	var calls int32
	fn := func(_ ...any) { atomic.AddInt32(&calls, 1) }

	delay := time.Second // long delay to make the window obvious
	throttled := New(fn, delay)

	const goroutines = 100
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			throttled() // all goroutines race to call at (roughly) the same time
		}()
	}
	wg.Wait()

	// give the throttled wrapper a short moment to invoke fn
	eventually(t, 20*time.Millisecond, func() bool { return atomic.LoadInt32(&calls) > 0 })

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("thread-safety check failed: expected 1 underlying call, got %d", got)
	}
}
