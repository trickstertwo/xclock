package xclock

import (
	"sync/atomic"
	"time"
)

// Singleton: global default clock with atomic access.

var (
	defaultClock atomic.Value // holds Clock
	system       atomic.Bool  // true when Default() is the standard system clock
)

func init() {
	defaultClock.Store(standardSystemClock)
	system.Store(true) // we start with the system clock
}

// Default returns the process-wide default Clock.
// Note: On hot paths, prefer the top-level Facade functions (Now/Sleep/etc.) which
// automatically fast-path the system clock, or capture Default() once at the call site.
func Default() Clock {
	c := defaultClock.Load()
	if c == nil {
		return standardSystemClock
	}
	return c.(Clock)
}

// SetDefault replaces the process-wide default Clock.
// Safe for concurrent readers. Prefer using only in main/init/tests.
func SetDefault(c Clock) {
	if c == nil {
		panic("xclock: SetDefault with nil Clock")
	}
	defaultClock.Store(c)
	// Update fast-path flag. Pointer equality is intentional: we only fast-path
	// when the exact singleton standardSystemClock is active.
	system.Store(c == standardSystemClock)
}

// NewFrozen returns a frozen clock at t.
func NewFrozen(t time.Time) Clock { return &frozenClock{t: t} }
