package xclock

import (
	"context"
	"sync/atomic"
	"time"
)

// Strategy: CalibratedClock applies a dynamic offset (delta) to time.Now(),
// typically learned from an external authority (NTP, secure time, GPS).
// - No background goroutines by default.
// - Call SyncOnce to calibrate, or StartAutoSync to enable periodic updates.
type CalibratedClock struct {
	base  Clock
	delta atomic.Int64 // nanoseconds to add to base.Now()
}

func NewCalibrated(base Clock) *CalibratedClock {
	if base == nil {
		base = standardSystemClock
	}
	return &CalibratedClock{base: base}
}

func (c *CalibratedClock) Now() time.Time {
	d := time.Duration(c.delta.Load())
	return c.base.Now().Add(d)
}

func (c *CalibratedClock) Since(t time.Time) time.Duration { return c.Now().Sub(t) }
func (c *CalibratedClock) Sleep(d time.Duration)           { c.base.Sleep(d) }
func (c *CalibratedClock) After(d time.Duration) <-chan time.Time {
	return c.base.After(d)
}
func (c *CalibratedClock) AfterFunc(d time.Duration, f func()) CancelFunc {
	return c.base.AfterFunc(d, f)
}
func (c *CalibratedClock) NewTimer(d time.Duration) Timer   { return c.base.NewTimer(d) }
func (c *CalibratedClock) NewTicker(d time.Duration) Ticker { return c.base.NewTicker(d) }

// SetOffset sets the delta to apply to base time.
func (c *CalibratedClock) SetOffset(d time.Duration) {
	c.delta.Store(int64(d))
}

// AdjustOffset adds d to the existing delta.
func (c *CalibratedClock) AdjustOffset(d time.Duration) {
	c.delta.Add(int64(d))
}

// SyncOnce uses fetch(ctx) → authoritative time to compute delta.
// delta = authoritative - base.Now()
func (c *CalibratedClock) SyncOnce(ctx context.Context, fetch func(context.Context) (time.Time, error)) error {
	t, err := fetch(ctx)
	if err != nil {
		return err
	}
	now := c.base.Now()
	c.SetOffset(t.Sub(now))
	return nil
}

// StartAutoSync starts a periodic calibration loop. Returns a cancel function.
// Uses the provided Clock for scheduling (defaults to c.base).
func (c *CalibratedClock) StartAutoSync(ctx context.Context, period time.Duration, fetch func(context.Context) (time.Time, error), sched Clock) (cancel func()) {
	if sched == nil {
		sched = c.base
	}
	stop := make(chan struct{})
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
	return func() { close(stop) }
}
