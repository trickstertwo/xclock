package xclock

import "time"

// Builder: construct clocks step-by-step using functional options.

type StrategyKind int

const (
	StrategySystem StrategyKind = iota
	StrategyFrozen
)

type Options struct {
	Strategy   StrategyKind
	FrozenTime time.Time
	Offset     time.Duration
	Jitter     time.Duration
	JitterSeed uint64
}

type Option func(*Options)

func WithStrategy(s StrategyKind) Option { return func(o *Options) { o.Strategy = s } }
func WithFrozenTime(t time.Time) Option  { return func(o *Options) { o.FrozenTime = t } }
func WithOffset(d time.Duration) Option  { return func(o *Options) { o.Offset = d } }
func WithJitter(d time.Duration) Option  { return func(o *Options) { o.Jitter = d } }
func WithJitterSeed(seed uint64) Option  { return func(o *Options) { o.JitterSeed = seed } }

type Builder struct {
	opts Options
}

func NewBuilder() *Builder {
	return &Builder{
		opts: Options{
			Strategy: StrategySystem,
		},
	}
}

func (b *Builder) Apply(opts ...Option) *Builder {
	for _, f := range opts {
		f(&b.opts)
	}
	return b
}

// Build constructs the Clock according to options (composition-friendly).
func (b *Builder) Build() Clock {
	var base Clock
	switch b.opts.Strategy {
	case StrategyFrozen:
		t := b.opts.FrozenTime
		if t.IsZero() {
			t = time.Unix(0, 0).UTC()
		}
		base = &frozenClock{t: t}
	case StrategySystem:
		fallthrough
	default:
		base = standardSystemClock
	}
	if b.opts.Offset != 0 {
		base = NewOffset(base, b.opts.Offset)
	}
	if b.opts.Jitter != 0 {
		if b.opts.JitterSeed != 0 {
			base = newJitterWithSeed(base, b.opts.Jitter, b.opts.JitterSeed)
		} else {
			base = NewJitter(base, b.opts.Jitter)
		}
	}
	return base
}
