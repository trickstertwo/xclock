package xclock

import (
	"sync/atomic"
	"time"
)

// facadeFns holds pre-bound function pointers for the facade.
// We atomically swap a pointer to this struct in SetDefault, so facade calls
// are narrowly on the hot path: a single atomic load + direct function call.
type facadeFns struct {
	now       func() time.Time
	since     func(time.Time) time.Duration
	sleep     func(time.Duration)
	after     func(time.Duration) <-chan time.Time
	afterFunc func(time.Duration, func()) CancelFunc
	newTimer  func(time.Duration) Timer
	newTicker func(time.Duration) Ticker
}

var fns atomic.Pointer[facadeFns]

// initFacadeFns sets fast-path to stdlib (system clock) without requiring adapters.
func initFacadeFns() {
	sys := &facadeFns{
		now:   time.Now,
		since: time.Since,
		sleep: time.Sleep,
		after: time.After,
		afterFunc: func(d time.Duration, f func()) CancelFunc {
			t := time.AfterFunc(d, f)
			return t.Stop
		},
		newTimer:  func(d time.Duration) Timer { return &stdTimer{t: time.NewTimer(d)} },
		newTicker: func(d time.Duration) Ticker { return &stdTicker{t: time.NewTicker(d)} },
	}
	fns.Store(sys)
}

// updateFacadeFns binds the facade to the provided Clock's methods.
func updateFacadeFns(c Clock) {
	g := &facadeFns{
		now:       c.Now,
		since:     c.Since,
		sleep:     c.Sleep,
		after:     c.After,
		afterFunc: c.AfterFunc,
		newTimer:  c.NewTimer,
		newTicker: c.NewTicker,
	}
	fns.Store(g)
}

// Facade: branchless calls via function pointers.

func Now() time.Time                  { return fns.Load().now() }
func Since(t time.Time) time.Duration { return fns.Load().since(t) }
func Sleep(d time.Duration)           { fns.Load().sleep(d) }
func After(d time.Duration) <-chan time.Time {
	return fns.Load().after(d)
}
func AfterFunc(d time.Duration, f func()) CancelFunc {
	return fns.Load().afterFunc(d, f)
}
func NewTimer(d time.Duration) Timer   { return fns.Load().newTimer(d) }
func NewTicker(d time.Duration) Ticker { return fns.Load().newTicker(d) }
