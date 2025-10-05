package xclock

import "time"

// Strategy: OffsetClock shifts the reported wall time by a fixed duration.
// Scheduling primitives (Sleep/After/Timer/Ticker) delegate to the base clock
// to avoid surprising real-time behavior changes.
type offsetClock struct {
	base   Clock
	offset time.Duration
}

func NewOffset(base Clock, d time.Duration) Clock {
	if base == nil {
		base = Default()
	}
	if d == 0 {
		return base
	}
	return &offsetClock{base: base, offset: d}
}

func (o *offsetClock) Now() time.Time                  { return o.base.Now().Add(o.offset) }
func (o *offsetClock) Since(t time.Time) time.Duration { return o.Now().Sub(t) }
func (o *offsetClock) Sleep(d time.Duration)           { o.base.Sleep(d) }
func (o *offsetClock) After(d time.Duration) <-chan time.Time {
	return o.base.After(d)
}
func (o *offsetClock) AfterFunc(d time.Duration, f func()) CancelFunc {
	return o.base.AfterFunc(d, f)
}
func (o *offsetClock) NewTimer(d time.Duration) Timer   { return o.base.NewTimer(d) }
func (o *offsetClock) NewTicker(d time.Duration) Ticker { return o.base.NewTicker(d) }
