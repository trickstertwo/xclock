package jitter

import (
	"sync/atomic"
	"time"

	"github.com/trickstertwo/xclock"
)

// Jitter clock: adds symmetric random jitter in [-MaxJitter, +MaxJitter]
// to observed wall time returned by Now(). Scheduling delegates to the base clock.
//
// Notes:
// - Only Now() is jittered; Sleep/After/Timers/Tickers use the base clock.
// - RNG is a lock-free SplitMix64 step on each Now() via atomic.Uint64.
// - MaxJitter <= 0 disables jitter (returns base as-is).

type Config struct {
	// Base is the underlying clock to wrap. If nil, xclock.Default() is used.
	Base xclock.Clock
	// MaxJitter is the maximum absolute jitter applied to Now().
	// The actual jitter is uniformly distributed in [-MaxJitter, +MaxJitter].
	MaxJitter time.Duration
	// Seed initializes the PRNG. If 0, it's seeded from time.Now().UnixNano().
	Seed uint64
}

// Set sets the jittered clock as the process-wide default and returns a restore
// function that reverts to the previous default when called.
// Recommended for tests and examples:
//
//	restore := jitter.Use(jitter.Config{MaxJitter: 2 * time.Millisecond})
//	defer restore()
func Set(cfg Config) (restore func()) {
	prev := xclock.Default()
	xclock.SetDefault(NewWithSeed(cfg.Base, cfg.MaxJitter, cfg.Seed))
	return func() { xclock.SetDefault(prev) }
}

// Use applies the jittered clock without returning a restore function.
// Recommended in production mains where you never intend to restore.
func Use(cfg Config) {
	xclock.SetDefault(NewWithSeed(cfg.Base, cfg.MaxJitter, cfg.Seed))
}

// With runs fn with the jittered clock active, then restores the previous clock
// even if fn panics (restore still runs during unwinding).
func With(cfg Config, fn func()) {
	restore := Set(cfg)
	defer restore()
	fn()
}

// New creates a jitter clock with a time-based seed.
func New(base xclock.Clock, maxJitter time.Duration) xclock.Clock {
	return NewWithSeed(base, maxJitter, 0)
}

// NewWithSeed constructs a jitter clock using the provided seed.
// If maxJitter <= 0, the base clock is returned directly (no jitter).
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
	seed      atomic.Uint64 // SplitMix64 state
}

func (j *clock) Now() time.Time {
	baseNow := j.base.Now()

	span := int64(j.maxJitter)
	if span <= 0 {
		return baseNow
	}

	// SplitMix64: increment then mix
	x := j.seed.Add(0x9e3779b97f4a7c15)
	z := x
	z ^= z >> 30
	z *= 0xbf58476d1ce4e5b9
	z ^= z >> 27
	z *= 0x94d049bb133111eb
	z ^= z >> 31

	// Uniform in [-span, +span]
	// Convert span to uint64 safely
	uSpan := uint64(span)
	jit := int64(z%(2*uSpan+1)) - span

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
