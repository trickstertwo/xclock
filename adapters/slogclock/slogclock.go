package slogclock

import (
	"context"
	"log/slog"

	"github.com/trickstertwo/xclock"
)

// Adapter: slog Handler wrapper that overrides Record time using xclock.Clock.
//
// Usage:
//
//	clk := xclock.Default()
//	h := slogclock.WithClock(slog.NewJSONHandler(os.Stdout, nil), clk)
//	logger := slog.New(h)
type handler struct {
	inner slog.Handler
	clk   xclock.Clock
}

func WithClock(h slog.Handler, c xclock.Clock) slog.Handler {
	if c == nil {
		c = xclock.Default()
	}
	return &handler{inner: h, clk: c}
}

func (h *handler) Enabled(ctx context.Context, lvl slog.Level) bool {
	return h.inner.Enabled(ctx, lvl)
}

func (h *handler) Handle(ctx context.Context, r slog.Record) error {
	r.Time = h.clk.Now()
	return h.inner.Handle(ctx, r)
}

func (h *handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &handler{inner: h.inner.WithAttrs(attrs), clk: h.clk}
}

func (h *handler) WithGroup(name string) slog.Handler {
	return &handler{inner: h.inner.WithGroup(name), clk: h.clk}
}
