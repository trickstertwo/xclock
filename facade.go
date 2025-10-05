package xclock

import "time"

// Facade: simplified global API with a fast-path.
// When the default clock is the standard system clock, we avoid the atomic.Value load
// and interface dispatch entirely by calling the stdlib directly. This is controlled
// by system, which is updated in SetDefault.

func Now() time.Time {
	if system.Load() {
		return time.Now()
	}
	return Default().Now()
}

func Since(t time.Time) time.Duration {
	if system.Load() {
		return time.Since(t)
	}
	return Default().Since(t)
}

func Sleep(d time.Duration) {
	if system.Load() {
		time.Sleep(d)
		return
	}
	Default().Sleep(d)
}

func After(d time.Duration) <-chan time.Time {
	if system.Load() {
		return time.After(d)
	}
	return Default().After(d)
}

func AfterFunc(d time.Duration, f func()) CancelFunc {
	if system.Load() {
		t := time.AfterFunc(d, f)
		return t.Stop
	}
	return Default().AfterFunc(d, f)
}

func NewTimer(d time.Duration) Timer {
	if system.Load() {
		return &stdTimer{t: time.NewTimer(d)}
	}
	return Default().NewTimer(d)
}

func NewTicker(d time.Duration) Ticker {
	if system.Load() {
		return &stdTicker{t: time.NewTicker(d)}
	}
	return Default().NewTicker(d)
}
