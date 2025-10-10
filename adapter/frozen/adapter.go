package frozen

import (
	"time"

	"github.com/trickstertwo/xclock"
)

// frozen adapter: deterministic Now(). Scheduling uses stdlib to avoid deadlocks.

type Config struct {
	Time time.Time
}

func Use(cfg Config) {
	xclock.SetDefault(New(cfg.Time))
}

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
