package jitter

import (
	"sync/atomic"
	"time"

	"github.com/trickstertwo/xclock"
)

// jitter adapter: adds symmetric random jitter in [-maxJitter, +maxJitter]
// to observed wall time. Scheduling delegates to base.

type Config struct {
	Base      xclock.Clock
	MaxJitter time.Duration
	Seed      uint64 // optional; if 0, seeded from time
}

func Use(cfg Config) {
	xclock.SetDefault(NewWithSeed(cfg.Base, cfg.MaxJitter, cfg.Seed))
}

func New(base xclock.Clock, maxJitter time.Duration) xclock.Clock {
	return NewWithSeed(base, maxJitter, 0)
}

func NewWithSeed(base xclock.Clock, maxJitter time.Duration, seed uint64) xclock.Clock {
	if base == nil {
		base = xclock.Default()
	}
	if maxJitter <= 0 {
		return base
	}
	j := &clock{base: base, maxJitter: maxJitter}
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	j.seed.Store(seed)
	return j
}

type clock struct {
	base      xclock.Clock
	maxJitter time.Duration
	seed      atomic.Uint64
}

func (j *clock) Now() time.Time {
	baseNow := j.base.Now()
	// SplitMix64 step
	x := j.seed.Add(0x9e3779b97f4a7c15)
	z := x
	z ^= z >> 30
	z *= 0xbf58476d1ce4e5b9
	z ^= z >> 27
	z *= 0x94d049bb133111eb
	z ^= z >> 31

	span := uint64(j.maxJitter.Nanoseconds())
	if span == 0 {
		return baseNow
	}
	// Range [-span, +span]
	jit := int64(z%(2*span+1)) - int64(span)
	return baseNow.Add(time.Duration(jit))
}

func (j *clock) Since(t time.Time) time.Duration { return j.Now().Sub(t) }
func (j *clock) Sleep(d time.Duration)           { j.base.Sleep(d) }
func (j *clock) After(d time.Duration) <-chan time.Time {
	return j.base.After(d)
}
func (j *clock) AfterFunc(d time.Duration, f func()) xclock.CancelFunc {
	return j.base.AfterFunc(d, f)
}
func (j *clock) NewTimer(d time.Duration) xclock.Timer   { return j.base.NewTimer(d) }
func (j *clock) NewTicker(d time.Duration) xclock.Ticker { return j.base.NewTicker(d) }
