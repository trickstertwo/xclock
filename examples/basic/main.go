package main

import (
	"fmt"
	"time"

	"github.com/trickstertwo/xclock"
)

func main() {
	// Fast-path active by default (system clock), no atomic loads taken.
	start := xclock.Now()
	xclock.Sleep(5 * time.Millisecond)
	elapsed := xclock.Since(start)
	fmt.Println("elapsed:", elapsed)

	// Capturing the Clock once avoids repeated lookups in hot paths.
	c := xclock.Default()
	t := c.NewTimer(10 * time.Millisecond)
	defer t.Stop()
	<-t.C()
	fmt.Println("timer fired")

	// Swapping default clock (e.g., in tests)
	xclock.SetDefault(xclock.NewFrozen(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)))
	fmt.Println("frozen now:", xclock.Now())
	// Swap back to system
	xclock.SetDefault(xclock.Default()) // no-op if already default system in your app
}
