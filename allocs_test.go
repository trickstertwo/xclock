package xclock

import (
	"testing"
)

// Verifies no allocations on the hot path via the facade.
func TestNowFacade_NoAllocs(t *testing.T) {
	// Warm up to avoid first-call effects.
	_ = Now()

	allocs := testing.AllocsPerRun(10000, func() {
		_ = Now()
	})
	if allocs != 0 {
		t.Fatalf("expected 0 allocs for Now facade, got %v", allocs)
	}
}

// Verifies no allocations when using a captured/injected Clock.
func TestNowInjected_NoAllocs(t *testing.T) {
	c := Default()
	// Warm up
	_ = c.Now()

	allocs := testing.AllocsPerRun(10000, func() {
		_ = c.Now()
	})
	if allocs != 0 {
		t.Fatalf("expected 0 allocs for injected Clock.Now, got %v", allocs)
	}
}
