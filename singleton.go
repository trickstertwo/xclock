package xclock

import (
	"sync/atomic"
)

// We store a stable wrapper type in atomic.Value to avoid type-mismatch panics.
type clockValue struct {
	c Clock
}

var (
	defaultClock atomic.Value // holds clockValue
)

func init() {
	// Initialize facade to stdlib fast-path and Default() to std clock.
	initFacadeFns()
	defaultClock.Store(clockValue{c: standardSystemClock})
}

// Default returns the process-wide default Clock.
func Default() Clock {
	v := defaultClock.Load()
	if v == nil {
		return standardSystemClock
	}
	return v.(clockValue).c
}

// SetDefault replaces the process-wide default Clock and rebinds the facade.
func SetDefault(c Clock) {
	if c == nil {
		panic("xclock: SetDefault with nil Clock")
	}
	defaultClock.Store(clockValue{c: c})
	updateFacadeFns(c)
}
