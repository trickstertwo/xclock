package main

import (
	"context"
	"fmt"
	"time"

	"github.com/trickstertwo/xclock"
	"github.com/trickstertwo/xclock/adapter/calibrated"
	"github.com/trickstertwo/xclock/adapter/compose"
)

func main() {
	old := xclock.Default()
	defer xclock.SetDefault(old)

	// Start from system time as baseline, mirroring the single Use(...) pattern.
	compose.Use(compose.Config{
		Strategy: compose.StrategySystem,
	})

	// Build a calibrated clock that learns an offset from an authority.
	c := calibrated.New(nil) // base = xclock.Default() (system)
	// One-shot sync: pretend authority is ahead by +500ms
	_ = c.SyncOnce(context.Background(), func(ctx context.Context) (time.Time, error) {
		return time.Now().Add(500 * time.Millisecond), nil
	})
	xclock.SetDefault(c)

	fmt.Println("== calibrated example ==")
	fmt.Println("Now() ~ base+500ms:", xclock.Now().Format(time.RFC3339Nano))

	// Auto-sync: adjust by +200ms more every 50ms (demo only).
	ctx, cancel := context.WithTimeout(context.Background(), 160*time.Millisecond)
	defer cancel()
	stop := c.StartAutoSync(ctx, 50*time.Millisecond, func(ctx context.Context) (time.Time, error) {
		// Simulate drift in source: each poll says "you're +200ms behind"
		return time.Now().Add(700 * time.Millisecond), nil // net +200ms over prior 500ms
	}, nil)
	defer stop()

	time.Sleep(120 * time.Millisecond)
	fmt.Println("Now() after auto-sync:", xclock.Now().Format(time.RFC3339Nano))
}
