package xclock

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestSleepContext_Cancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := SleepContext(ctx, 50*time.Millisecond, nil)
	if err == nil {
		t.Fatal("expected context error on canceled context")
	}
}

func TestSleepContext_Deadline(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := SleepContext(ctx, 15*time.Millisecond, nil)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should be at least requested sleep
	if elapsed < 15*time.Millisecond {
		t.Fatalf("slept too short: %v", elapsed)
	}
}

func TestUntil_CallsFunction(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	var calls atomic.Int32
	Until(ctx, 25*time.Millisecond, func(_ time.Time) {
		calls.Add(1)
	}, nil)

	if calls.Load() == 0 {
		t.Fatal("Until did not invoke function")
	}
}
