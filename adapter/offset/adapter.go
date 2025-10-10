package offset

import (
	"sync/atomic"
	"time"

	"github.com/trickstertwo/xclock"
)

// Offset clock: shifts observed wall time by a fixed duration.
// Scheduling delegates to the base clock unchanged.
//
// Notes:
// - Only Now/Since use the offset; Sleep/After/Timers/Tickers are delegated.
// - The offset can be adjusted at runtime via SetOffset/AdjustOffset.

type Config struct {
	// Base is the underlying clock to wrap. If nil, xclock.Default() is used.
	Base xclock.Clock
	// Offset is the amount of time to add to base.Now().
	Offset time.Duration
}

// Set sets the offset clock as the process-wide default and returns a restore
// function that reverts to the previous default when called.
// Recommended for tests and examples:
//
//	restore := offset.Use(offset.Config{Offset: 500 * time.Millisecond})
//	defer restore()
func Set(cfg Config) (restore func()) {
	prev := xclock.Default()
	xclock.SetDefault(New(cfg.Base, cfg.Offset))
	return func() { xclock.SetDefault(prev) }
}

// Use applies the offset clock without returning a restore function.
// Recommended in production mains where you never intend to restore.
func Use(cfg Config) {
	xclock.SetDefault(New(cfg.Base, cfg.Offset))
}

// With runs fn with the offset clock active, then restores the previous clock
// even if fn panics (restore still runs during unwinding).
func With(cfg Config, fn func()) {
	restore := Set(cfg)
	defer restore()
	fn()
}

// New constructs an offset clock. If base is nil, xclock.Default() is used.
// If d == 0, the base clock is returned directly.
func New(base xclock.Clock, d time.Duration) xclock.Clock {
	if base == nil {
		base = xclock.Default()
	}
	if d == 0 {
		return base
	}
	c := &clock{base: base}
	c.offset.Store(int64(d))
	return c
}

type clock struct {
	base   xclock.Clock
	offset atomic.Int64 // nanoseconds
}

func (o *clock) Now() time.Time {
	off := time.Duration(o.offset.Load())
	return o.base.Now().Add(off)
}

func (o *clock) Since(t time.Time) time.Duration { return o.Now().Sub(t) }
func (o *clock) Sleep(d time.Duration)           { o.base.Sleep(d) }
func (o *clock) After(d time.Duration) <-chan time.Time {
	return o.base.After(d)
}
func (o *clock) AfterFunc(d time.Duration, f func()) xclock.CancelFunc {
	return o.base.AfterFunc(d, f)
}
func (o *clock) NewTimer(d time.Duration) xclock.Timer   { return o.base.NewTimer(d) }
func (o *clock) NewTicker(d time.Duration) xclock.Ticker { return o.base.NewTicker(d) }

// SetOffset sets the absolute offset applied to base time.
func (o *clock) SetOffset(d time.Duration) { o.offset.Store(int64(d)) }

// AdjustOffset adds d to the current offset (can be negative).
func (o *clock) AdjustOffset(d time.Duration) { o.offset.Add(int64(d)) }

// Offset returns the current configured offset.
func (o *clock) Offset() time.Duration { return time.Duration(o.offset.Load()) }
