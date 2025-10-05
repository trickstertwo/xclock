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
}

type Option func(*Options)

// WithStrategy sets the clock strategy.
func WithStrategy(s StrategyKind) Option {
	return func(o *Options) { o.Strategy = s }
}

// WithFrozenTime sets the frozen time for StrategyFrozen.
func WithFrozenTime(t time.Time) Option {
	return func(o *Options) { o.FrozenTime = t }
}

type Builder struct {
	opts Options
}

// NewBuilder creates a new clock builder with sane defaults.
func NewBuilder() *Builder {
	return &Builder{
		opts: Options{
			Strategy: StrategySystem,
		},
	}
}

// Apply options.
func (b *Builder) Apply(opts ...Option) *Builder {
	for _, f := range opts {
		f(&b.opts)
	}
	return b
}

// Build constructs the Clock according to options.
func (b *Builder) Build() Clock {
	switch b.opts.Strategy {
	case StrategyFrozen:
		t := b.opts.FrozenTime
		if t.IsZero() {
			t = time.Unix(0, 0).UTC()
		}
		return &frozenClock{t: t}
	case StrategySystem:
		fallthrough
	default:
		return standardSystemClock
	}
}
