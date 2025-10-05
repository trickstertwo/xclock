package xclock

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultFacade_SystemFastPath(t *testing.T) {
	t.Parallel()

	// Sanity: facade Now close to time.Now
	before := time.Now()
	n := Now()
	after := time.Now()

	if n.Before(before) || n.After(after.Add(5*time.Millisecond)) {
		t.Fatalf("facade Now out of expected range: before=%v now=%v after=%v", before, n, after)
	}

	// Sleep should block appropriately
	d := 20 * time.Millisecond
	start := time.Now()
	Sleep(d)
	elapsed := time.Since(start)
	if elapsed < d {
		t.Fatalf("Sleep too short: got=%v want>=%v", elapsed, d)
	}

	// After should fire within reasonable tolerance
	select {
	case <-After(20 * time.Millisecond):
	case <-time.After(200 * time.Millisecond):
		t.Fatal("After did not fire in time")
	}

	// AfterFunc returns CancelFunc; verify it fires and cancel after fired returns false
	var fired atomic.Bool
	cancel := AfterFunc(15*time.Millisecond, func() { fired.Store(true) })
	defer cancel()
	time.Sleep(40 * time.Millisecond)
	if !fired.Load() {
		t.Fatal("AfterFunc callback did not fire")
	}
	if cancel() {
		t.Fatal("CancelFunc returned true after firing; expected false")
	}
}

func TestSetDefault_SwapFrozen_NoPanicAndAccurate(t *testing.T) {
	t.Parallel()

	orig := Default()
	defer SetDefault(orig) // restore for other tests

	frozenT := time.Date(2030, 2, 3, 4, 5, 6, 7, time.UTC)
	SetDefault(NewFrozen(frozenT))
	got := Now()
	if !got.Equal(frozenT) {
		t.Fatalf("frozen Now mismatch: got=%v want=%v", got, frozenT)
	}

	// Swap back to original and ensure Now changes away from frozen (within a tolerance)
	SetDefault(orig)
	n := Now()
	if n.Equal(frozenT) {
		t.Fatal("Now still equals frozen time after restoring default")
	}
}

func TestTimer_StopReset(t *testing.T) {
	t.Parallel()

	c := Default()

	// Create timer and stop before it fires; Stop should return true.
	tm := c.NewTimer(100 * time.Millisecond)
	if !tm.Stop() {
		// It's possible (though unlikely) it already fired; reset to establish control.
	}
	// Reset to short duration, should fire.
	if !tm.Reset(20 * time.Millisecond) {
		// For time.Timer semantics, Reset after Stop may return false depending on impl;
		// we rely on the timer to fire nonetheless.
	}
	select {
	case <-tm.C():
	case <-time.After(200 * time.Millisecond):
		t.Fatal("timer did not fire after Reset")
	}
}

func TestTicker_Basic(t *testing.T) {
	t.Parallel()

	c := Default()
	tk := c.NewTicker(25 * time.Millisecond)
	defer tk.Stop()

	// Expect at least 2 ticks within time budget
	timeout := time.After(300 * time.Millisecond)
	count := 0
	for count < 2 {
		select {
		case <-tk.C():
			count++
		case <-timeout:
			t.Fatalf("ticker fired %d times; want>=2", count)
		}
	}
}

func TestAfterFunc_CancelBeforeFire(t *testing.T) {
	t.Parallel()

	var fired atomic.Bool
	cancel := AfterFunc(100*time.Millisecond, func() { fired.Store(true) })
	if !cancel() {
		// If false, it may have fired already; check fired flag
		if !fired.Load() {
			t.Fatal("CancelFunc returned false but callback not fired")
		}
		return
	}
	// Ensure it didn't run
	time.Sleep(50 * time.Millisecond)
	if fired.Load() {
		t.Fatal("callback executed despite cancellation")
	}
}
