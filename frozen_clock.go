package xclock

import "time"

// Strategy: Frozen clock for deterministic tests. Not intended for production.
// Sleep/After/etc. still use stdlib to avoid surprising deadlocks in prod if misused.

type frozenClock struct {
	t time.Time
}

func (f *frozenClock) Now() time.Time                  { return f.t }
func (f *frozenClock) Since(t time.Time) time.Duration { return f.t.Sub(t) }
func (f *frozenClock) Sleep(d time.Duration)           { time.Sleep(d) } // intentionally real sleep
func (f *frozenClock) After(d time.Duration) <-chan time.Time {
	return standardSystemClock.After(d)
}
func (f *frozenClock) AfterFunc(d time.Duration, fn func()) CancelFunc {
	return standardSystemClock.AfterFunc(d, fn)
}
func (f *frozenClock) NewTimer(d time.Duration) Timer   { return standardSystemClock.NewTimer(d) }
func (f *frozenClock) NewTicker(d time.Duration) Ticker { return standardSystemClock.NewTicker(d) }
