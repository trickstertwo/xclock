package zapclock

import (
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/trickstertwo/xclock"
)

// Adapter: zapcore.Clock implementation backed by xclock.Clock.
//
// Usage:
//
//	clk := xclock.Default()
//	logger := zap.New(core, zap.WithClock(zapclock.New(clk)))
type clockAdapter struct {
	c xclock.Clock
}

func New(c xclock.Clock) zapcore.Clock {
	if c == nil {
		c = xclock.Default()
	}
	return &clockAdapter{c: c}
}

func (z *clockAdapter) Now() time.Time { return z.c.Now() }

// Zap's Clock also has NewTicker returning *time.Ticker; we provide stdlib ticker.
func (z *clockAdapter) NewTicker(d time.Duration) *time.Ticker {
	return time.NewTicker(d)
}
