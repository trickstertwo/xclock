package main

import (
	"fmt"
	"time"

	"github.com/trickstertwo/xclock"
	"github.com/trickstertwo/xclock/adapter/compose"
)

func main() {
	// Preserve and restore previous default to keep examples isolated.
	old := xclock.Default()
	defer xclock.SetDefault(old)

	// Optimal: configure via a single explicit Use(...) call.
	// - Base: system or frozen (deterministic demo here).
	// - Offset: shift reported Now() by +2s.
	// - Jitter: 0 (set if you want symmetric random jitter).
	compose.Use(compose.Config{
		Strategy:   compose.StrategyFrozen,
		FrozenTime: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Offset:     2 * time.Second,
		// Jitter:     2 * time.Millisecond,
		// JitterSeed: 42,
	})

	fmt.Println("== compose example ==")
	fmt.Println("Now()", xclock.Now().Format(time.RFC3339Nano))

	// Facade timers/tickers work with any adapter.
	t := xclock.NewTimer(15 * time.Millisecond)
	defer t.Stop()

	select {
	case tt := <-t.C():
		fmt.Println("Timer fired at", tt.Format(time.RFC3339Nano))
	case <-time.After(200 * time.Millisecond):
		fmt.Println("Timer did not fire in time")
	}

	tk := xclock.NewTicker(20 * time.Millisecond)
	defer tk.Stop()

	// Expect 2 quick ticks for demo brevity.
	for i := 0; i < 2; i++ {
		tt := <-tk.C()
		fmt.Println("Tick", i+1, "at", tt.Format(time.RFC3339Nano))
	}
}
