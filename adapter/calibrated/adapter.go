package calibrated

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trickstertwo/xclock"
)

// Calibrated clock: applies a dynamic offset (delta) to base.Now().
// Typical use: learn delta from an external authority (NTP, secure time, GPS).
// No background goroutines unless StartAutoSync is used.

type Config struct {
	// Base is the underlying clock to calibrate. If nil, xclock.Default() is used.
	Base xclock.Clock
	// InitialOffset is an optional offset to apply immediately.
	InitialOffset time.Duration
}

// Set sets the calibrated clock as the process-wide default and returns a restore
// function that reverts to the previous default when called.
// Recommended for tests and examples:
//
//	restore := calibrated.Use(calibrated.Config{Base: xclock.System(), InitialOffset: 50 * time.Millisecond})
//	defer restore()
func Set(cfg Config) (restore func()) {
	prev := xclock.Default()
	c := New(cfg.Base)
	if cfg.InitialOffset != 0 {
		c.SetOffset(cfg.InitialOffset)
	}
	xclock.SetDefault(c)
	return func() { xclock.SetDefault(prev) }
}

// Use applies the calibrated clock without returning a restore function.
// Recommended in production mains where you never intend to restore.
func Use(cfg Config) {
	c := New(cfg.Base)
	if cfg.InitialOffset != 0 {
		c.SetOffset(cfg.InitialOffset)
	}
	xclock.SetDefault(c)
}

// With runs fn with the calibrated clock active, then restores the previous clock
// even if fn panics (restore still runs during unwinding).
func With(cfg Config, fn func()) {
	restore := Set(cfg)
	defer restore()
	fn()
}

type Clock struct {
	base  xclock.Clock
	delta atomic.Int64 // nanoseconds to add to base.Now()
}

func New(base xclock.Clock) *Clock {
	if base == nil {
		base = xclock.Default()
	}
	return &Clock{base: base}
}

// Now returns base.Now() + current offset (delta).
func (c *Clock) Now() time.Time {
	d := time.Duration(c.delta.Load())
	return c.base.Now().Add(d)
}

func (c *Clock) Since(t time.Time) time.Duration { return c.Now().Sub(t) }
func (c *Clock) Sleep(d time.Duration)           { c.base.Sleep(d) }
func (c *Clock) After(d time.Duration) <-chan time.Time {
	return c.base.After(d)
}
func (c *Clock) AfterFunc(d time.Duration, f func()) xclock.CancelFunc {
	return c.base.AfterFunc(d, f)
}
func (c *Clock) NewTimer(d time.Duration) xclock.Timer   { return c.base.NewTimer(d) }
func (c *Clock) NewTicker(d time.Duration) xclock.Ticker { return c.base.NewTicker(d) }

// SetOffset sets the absolute delta to apply to base time.
func (c *Clock) SetOffset(d time.Duration) { c.delta.Store(int64(d)) }

// AdjustOffset adds d to the existing delta.
func (c *Clock) AdjustOffset(d time.Duration) { c.delta.Add(int64(d)) }

// Offset returns the current delta.
func (c *Clock) Offset() time.Duration { return time.Duration(c.delta.Load()) }

// SyncOnce uses fetch(ctx) → authoritative time to compute delta.
// delta = authoritative - base.Now(). Best-effort; leaves the previous delta on error.
func (c *Clock) SyncOnce(ctx context.Context, fetch func(context.Context) (time.Time, error)) error {
	t, err := fetch(ctx)
	if err != nil {
		return err
	}
	now := c.base.Now()
	c.SetOffset(t.Sub(now))
	return nil
}

// StartAutoSync starts a periodic calibration loop. Returns a cancel function
// that is safe to call multiple times (idempotent). Uses the provided scheduler
// for ticks (defaults to c.base).
func (c *Clock) StartAutoSync(ctx context.Context, period time.Duration, fetch func(context.Context) (time.Time, error), sched xclock.Clock) (cancel func()) {
	if period <= 0 {
		// No-op cancel for invalid period, keeps API safe.
		return func() {}
	}
	if sched == nil {
		sched = c.base
	}
	stop := make(chan struct{})
	var once sync.Once

	go func() {
		tk := sched.NewTicker(period)
		defer tk.Stop()
		for {
			select {
			case <-tk.C():
				_ = c.SyncOnce(ctx, fetch) // best-effort
			case <-stop:
				return
			case <-ctx.Done():
				return
			}
		}
	}()

	return func() {
		once.Do(func() { close(stop) })
	}
}
