package xclock

import (
	"fmt"
	"sync"
)

// Factory: plugin registry for clock providers by name.

type Provider func(opts Options) (Clock, error)

var (
	regMu    sync.RWMutex
	registry = map[string]Provider{
		"system": func(Options) (Clock, error) { return standardSystemClock, nil },
		"frozen": func(o Options) (Clock, error) {
			return (&Builder{}).Apply(WithStrategy(StrategyFrozen), WithFrozenTime(o.FrozenTime)).Build(), nil
		},
		"offset": func(o Options) (Clock, error) {
			return (&Builder{}).Apply(WithOffset(o.Offset)).Build(), nil
		},
		"jitter": func(o Options) (Clock, error) {
			return (&Builder{}).Apply(WithJitter(o.Jitter), WithJitterSeed(o.JitterSeed)).Build(), nil
		},
		"compose": func(o Options) (Clock, error) {
			// Compose system/frozen + optional offset + jitter (with optional seed).
			return (&Builder{}).Apply(
				WithStrategy(o.Strategy),
				WithFrozenTime(o.FrozenTime),
				WithOffset(o.Offset),
				WithJitter(o.Jitter),
				WithJitterSeed(o.JitterSeed),
			).Build(), nil
		},
	}
)

// Register adds a named provider. It panics if name is already registered or empty.
func Register(name string, p Provider) {
	if name == "" {
		panic("xclock: Register requires non-empty name")
	}
	regMu.Lock()
	defer regMu.Unlock()
	if _, ok := registry[name]; ok {
		panic("xclock: duplicate provider: " + name)
	}
	registry[name] = p
}

// NewFromFactory constructs a clock by provider name.
func NewFromFactory(name string, opts Options) (Clock, error) {
	regMu.RLock()
	p, ok := registry[name]
	regMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("xclock: unknown provider %q", name)
	}
	return p(opts)
}
