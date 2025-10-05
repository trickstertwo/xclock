package slogclock

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/trickstertwo/xclock"
)

type captureHandler struct {
	last slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	// Make a copy of the record to persist beyond call
	r2 := slog.Record{
		Time:    r.Time,
		Message: r.Message,
		Level:   r.Level,
	}
	h.last = r2
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func TestSlogClock_UsesXclockTime(t *testing.T) {
	t.Parallel()

	ft := time.Date(2099, 12, 31, 23, 59, 59, 123, time.UTC)
	h := &captureHandler{}
	wrapped := WithClock(h, xclock.NewFrozen(ft))
	log := slog.New(wrapped)

	log.Info("test")
	if h.last.Time.IsZero() {
		t.Fatal("no record captured")
	}
	if !h.last.Time.Equal(ft) {
		t.Fatalf("record time mismatch: got=%v want=%v", h.last.Time, ft)
	}
}
