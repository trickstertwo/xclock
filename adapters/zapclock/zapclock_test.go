package zapclock

import (
	"testing"
	"time"

	"github.com/trickstertwo/xclock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestZapClock_UsesXclockTime(t *testing.T) {
	t.Parallel()

	ft := time.Date(2042, 3, 4, 5, 6, 7, 8, time.UTC)
	frozen := xclock.NewFrozen(ft)

	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core, zap.WithClock(New(frozen)))
	defer logger.Sync()

	logger.Info("hello")

	entries := obs.All()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	got := entries[0].Time
	if !got.Equal(ft) {
		t.Fatalf("zap entry time mismatch: got=%v want=%v", got, ft)
	}
}
