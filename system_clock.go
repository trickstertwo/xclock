package xclock

import "time"

// Strategy: System clock using stdlib time package.
// This is the dependable, high-performance default.

var standardSystemClock = &systemClock{}

type systemClock struct{}

func (s *systemClock) Now() time.Time                  { return time.Now() }
func (s *systemClock) Since(t time.Time) time.Duration { return time.Since(t) }
func (s *systemClock) Sleep(d time.Duration)           { time.Sleep(d) }
func (s *systemClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
func (s *systemClock) AfterFunc(d time.Duration, f func()) CancelFunc {
	t := time.AfterFunc(d, f)
	return t.Stop
}
func (s *systemClock) NewTimer(d time.Duration) Timer   { return &stdTimer{t: time.NewTimer(d)} }
func (s *systemClock) NewTicker(d time.Duration) Ticker { return &stdTicker{t: time.NewTicker(d)} }

// Adapter types to satisfy our interfaces with minimal overhead.

type stdTicker struct{ t *time.Ticker }

func (t *stdTicker) C() <-chan time.Time   { return t.t.C }
func (t *stdTicker) Stop()                 { t.t.Stop() }
func (t *stdTicker) Reset(d time.Duration) { t.t.Reset(d) }

type stdTimer struct{ t *time.Timer }

func (t *stdTimer) C() <-chan time.Time        { return t.t.C }
func (t *stdTimer) Stop() bool                 { return t.t.Stop() }
func (t *stdTimer) Reset(d time.Duration) bool { return t.t.Reset(d) }

// System returns the stdlib-backed system clock. Useful for adapters that want
// to explicitly select a system base without importing any adapter.
func System() Clock { return standardSystemClock }
