package xclock

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestFacade_RebindOnSetDefault(t *testing.T) {
	t.Parallel()

	orig := Default()
	t.Cleanup(func() { SetDefault(orig) })

	// Bind to system (stdlib fast-path)
	SetDefault(standardSystemClock)
	sys1 := Now()
	time.Sleep(2 * time.Millisecond)
	sys2 := Now()
	if !sys2.After(sys1) {
		t.Fatalf("system Now did not advance: %v -> %v", sys1, sys2)
	}

	// Swap to frozen and ensure facade reflects it
	ft := time.Date(2033, 5, 6, 7, 8, 9, 10, time.UTC)
	SetDefault(NewFrozen(ft))
	if got := Now(); !got.Equal(ft) {
		t.Fatalf("facade not rebound to frozen: got=%v want=%v", got, ft)
	}

	// Swap back to system and ensure it no longer equals frozen
	SetDefault(standardSystemClock)
	rest := Now()
	if rest.Equal(ft) {
		t.Fatalf("facade still returning frozen after restore: %v", rest)
	}
}

func TestConcurrent_SetDefaultAndNow_NoPanics(t *testing.T) {
	t.Parallel()

	orig := Default()
	t.Cleanup(func() { SetDefault(orig) })

	f1 := NewFrozen(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	f2 := NewFrozen(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup

	// Writers: flip defaults repeatedly
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(ix int) {
			defer wg.Done()
			for ctx.Err() == nil {
				if ix%2 == 0 {
					SetDefault(f1)
				} else {
					SetDefault(f2)
				}
			}
		}(i)
	}

	// Readers: hammer facade functions
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				_ = Now()
				_ = Since(Now())
				_ = AfterFunc(1*time.Nanosecond, func() {}) // minimal timer path
			}
		}()
	}

	wg.Wait()
	// If the test completes without panic or data race (run with -race), it's a pass.
}

func TestTicker_Reset(t *testing.T) {
	t.Parallel()

	c := Default()
	tk := c.NewTicker(50 * time.Millisecond)
	defer tk.Stop()

	// Wait for initial tick
	select {
	case <-tk.C():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("initial tick not received")
	}

	// Reset to a shorter period and ensure we see a faster tick
	start := time.Now()
	tk.Reset(10 * time.Millisecond)
	select {
	case <-tk.C():
	case <-time.After(100 * time.Millisecond):
		t.Fatal("no tick after Reset to shorter duration")
	}
	if elapsed := time.Since(start); elapsed > 80*time.Millisecond {
		t.Fatalf("Reset did not shorten interval: %v", elapsed)
	}
}
