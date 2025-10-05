package xclock

import (
	"sync/atomic"
	"time"
)

// Strategy: JitterClock adds symmetric random jitter in [-maxJitter, +maxJitter]
// to observed wall time. Scheduling primitives delegate to base.
type jitterClock struct {
	base      Clock
	maxJitter time.Duration
	seed      atomic.Uint64
}

func NewJitter(base Clock, maxJitter time.Duration) Clock {
	if base == nil {
		base = Default()
	}
	if maxJitter <= 0 {
		return base
	}
	j := &jitterClock{base: base, maxJitter: maxJitter}
	// Seed with current nanotime; exact value not critical.
	j.seed.Store(uint64(time.Now().UnixNano()))
	return j
}

// newJitterWithSeed allows deterministic jitter sequences (used via Builder).
func newJitterWithSeed(base Clock, maxJitter time.Duration, seed uint64) Clock {
	if base == nil {
		base = Default()
	}
	if maxJitter <= 0 {
		return base
	}
	j := &jitterClock{base: base, maxJitter: maxJitter}
	if seed == 0 {
		seed = uint64(time.Now().UnixNano())
	}
	j.seed.Store(seed)
	return j
}

func (j *jitterClock) Now() time.Time {
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

func (j *jitterClock) Since(t time.Time) time.Duration { return j.Now().Sub(t) }
func (j *jitterClock) Sleep(d time.Duration)           { j.base.Sleep(d) }
func (j *jitterClock) After(d time.Duration) <-chan time.Time {
	return j.base.After(d)
}
func (j *jitterClock) AfterFunc(d time.Duration, f func()) CancelFunc {
	return j.base.AfterFunc(d, f)
}
func (j *jitterClock) NewTimer(d time.Duration) Timer   { return j.base.NewTimer(d) }
func (j *jitterClock) NewTicker(d time.Duration) Ticker { return j.base.NewTicker(d) }
