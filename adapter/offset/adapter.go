package offset

import (
	"time"

	"github.com/trickstertwo/xclock"
)

// offset adapter: shifts observed wall time by a fixed duration.
// Scheduling delegates to the base clock unchanged.

type Config struct {
	Base   xclock.Clock
	Offset time.Duration
}

func Use(cfg Config) {
	xclock.SetDefault(New(cfg.Base, cfg.Offset))
}

func New(base xclock.Clock, d time.Duration) xclock.Clock {
	if base == nil {
		base = xclock.Default()
	}
	if d == 0 {
		return base
	}
	return &clock{base: base, offset: d}
}

type clock struct {
	base   xclock.Clock
	offset time.Duration
}

func (o *clock) Now() time.Time                  { return o.base.Now().Add(o.offset) }
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
