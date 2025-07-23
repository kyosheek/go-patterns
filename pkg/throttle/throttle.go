package throttle

import (
	"sync"
	"time"
)

// Fn is a side effect function with any number of arguments, and no return value.
type Fn func(args ...any)

// New returns a throttled version of fn.
// The throttled function calls fn immediately the first time
// and then ignores subsequent calls until the delay has elapsed.
func New(fn Fn, delay time.Duration) Fn {
	var (
		mu    sync.Mutex // protects timer
		timer *time.Timer
	)

	return func(args ...any) {
		mu.Lock()
		defer mu.Unlock()

		if timer == nil {
			fn(args...)

			// Start a one-shot timer that will reset the guard
			// after the specified delay.
			timer = time.AfterFunc(delay, func() {
				mu.Lock()
				timer = nil
				mu.Unlock()
			},
			)
		}
	}
}
