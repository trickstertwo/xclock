package compose

import (
	"time"

	"github.com/trickstertwo/xclock"
	"github.com/trickstertwo/xclock/adapter/frozen"
	"github.com/trickstertwo/xclock/adapter/jitter"
	"github.com/trickstertwo/xclock/adapter/offset"
)

// compose adapter: one-shot builder to set a composed default clock.
// Strategy + optional offset + optional jitter. Mirrors xlog.Use pattern.

type StrategyKind int

const (
	StrategySystem StrategyKind = iota
	StrategyFrozen
)

type Config struct {
	Strategy   StrategyKind
	FrozenTime time.Time
	Offset     time.Duration
	Jitter     time.Duration
	JitterSeed uint64
}

func Use(cfg Config) {
	// Base
	var base xclock.Clock
	switch cfg.Strategy {
	case StrategyFrozen:
		t := cfg.FrozenTime
		if t.IsZero() {
			t = time.Unix(0, 0).UTC()
		}
		base = frozen.New(t)
	case StrategySystem:
		fallthrough
	default:
		// Explicitly select the core system clock.
		base = xclock.System()
	}

	// Layers
	if cfg.Offset != 0 {
		base = offset.New(base, cfg.Offset)
	}
	if cfg.Jitter != 0 {
		if cfg.JitterSeed != 0 {
			base = jitter.NewWithSeed(base, cfg.Jitter, cfg.JitterSeed)
		} else {
			base = jitter.New(base, cfg.Jitter)
		}
	}

	xclock.SetDefault(base)
}
