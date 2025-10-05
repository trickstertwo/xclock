package main

import (
	"context"
	"fmt"
	"time"

	"github.com/trickstertwo/xclock"
)

func main() {
	// Your logger code remains unchanged; it reads xclock.Now() via adapters or DI.

	// Start with system time (default fast-path).
	fmt.Println("system:", xclock.Now())

	// Swap in offset (+50ms) without touching logger code.
	clkOff := xclock.NewBuilder().Apply(xclock.WithOffset(24 * time.Hour)).Build()
	xclock.SetDefault(clkOff)
	fmt.Println("offset+24h:", xclock.Now())

	// Compose offset (+50ms) + jitter (±5ms).
	clkJ := xclock.NewBuilder().Apply(
		xclock.WithOffset(24*time.Hour),
		xclock.WithJitter(12*time.Hour),
	).Build()
	xclock.SetDefault(clkJ)
	fmt.Println("offset+24h ±12h jitter:", xclock.Now())

	// Calibrated clock example (simulate NTP/secure time).
	cal := xclock.NewCalibrated(nil) // base = system
	// Simulate authoritative time being +10min ahead.
	_ = cal.SyncOnce(context.Background(), func(context.Context) (time.Time, error) {
		return time.Now().Add(1 * time.Minute), nil
	})
	xclock.SetDefault(cal)
	fmt.Println("calibrated (+1min):", xclock.Now())

	// Restore system.
	xclock.SetDefault(xclock.Default()) // if original was system, this is a no-op
}
