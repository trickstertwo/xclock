package frozen

import (
	"time"

	"github.com/trickstertwo/xclock"
)

// frozen adapter: deterministic Now(). Scheduling uses stdlib to avoid deadlocks.

type Config struct {
	Time time.Time
}

// Set sets the frozen clock as the process-wide default and returns a restore
// function that reverts to the previous default when called.
// Recommended in tests and examples:
//
//	restore := frozen.Use(frozen.Config{Time: t})
//	defer restore()
func Set(cfg Config) (restore func()) {
	prev := xclock.Default()
	xclock.SetDefault(New(cfg.Time))
	return func() { xclock.SetDefault(prev) }
}

// Use applies the frozen clock without returning a restore function.
// Recommended in production mains where you never intend to restore.
func Use(cfg Config) {
	xclock.SetDefault(New(cfg.Time))
}

// With runs fn with the frozen clock active, then restores the previous clock
// even if fn panics (restore still runs during unwinding).
func With(cfg Config, fn func()) {
	restore := Set(cfg)
	defer restore()
	fn()
}

// New constructs a frozen Clock instance at t.
func New(t time.Time) xclock.Clock {
	return &clock{t: t}
}

type clock struct {
	t time.Time
}

func (f *clock) Now() time.Time                  { return f.t }
func (f *clock) Since(t time.Time) time.Duration { return f.t.Sub(t) }
func (f *clock) Sleep(d time.Duration)           { time.Sleep(d) } // intentionally real
func (f *clock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}
func (f *clock) AfterFunc(d time.Duration, fn func()) xclock.CancelFunc {
	t := time.AfterFunc(d, fn)
	return t.Stop
}
func (f *clock) NewTimer(d time.Duration) xclock.Timer   { return &stdTimer{t: time.NewTimer(d)} }
func (f *clock) NewTicker(d time.Duration) xclock.Ticker { return &stdTicker{t: time.NewTicker(d)} }

type stdTicker struct{ t *time.Ticker }

func (t *stdTicker) C() <-chan time.Time   { return t.t.C }
func (t *stdTicker) Stop()                 { t.t.Stop() }
func (t *stdTicker) Reset(d time.Duration) { t.t.Reset(d) }

type stdTimer struct{ t *time.Timer }

func (t *stdTimer) C() <-chan time.Time        { return t.t.C }
func (t *stdTimer) Stop() bool                 { return t.t.Stop() }
func (t *stdTimer) Reset(d time.Duration) bool { return t.t.Reset(d) }
