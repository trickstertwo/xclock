package xclock

import (
	"time"
)

// CancelFunc mirrors time.Timer.Stop semantics when used by AfterFunc.
// It returns true if the timer was active before the call.
type CancelFunc func() bool

// Ticker abstracts time.Ticker.
type Ticker interface {
	C() <-chan time.Time
	Stop()
	Reset(d time.Duration)
}

// Timer abstracts time.Timer.
type Timer interface {
	C() <-chan time.Time
	Stop() bool
	Reset(d time.Duration) bool
}

// Clock is the Strategy interface for time access.
// Keep this interface minimal for stability and substitutability.
type Clock interface {
	// Now returns current time. Should be monotonic-safe where applicable.
	Now() time.Time
	// Since returns time elapsed since t.
	Since(t time.Time) time.Duration
	// Sleep blocks for d duration.
	Sleep(d time.Duration)
	// After returns a channel that fires once after d.
	After(d time.Duration) <-chan time.Time
	// AfterFunc runs f after d and returns a CancelFunc analogous to time.AfterFunc.
	AfterFunc(d time.Duration, f func()) CancelFunc
	// NewTimer returns a Timer that fires once after d.
	NewTimer(d time.Duration) Timer
	// NewTicker returns a Ticker that ticks every d.
	NewTicker(d time.Duration) Ticker
}
