package xclock

import (
	"testing"
	"time"
)

// Baseline: stdlib time.Now
func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now()
	}
}

// Facade fast-path (branchless, function pointers)
func BenchmarkNowFacade_System(b *testing.B) {
	orig := Default()
	defer SetDefault(orig)
	// Ensure system fast-path is bound (stdlib direct)
	SetDefault(standardSystemClock)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Now()
	}
}

// Injected clock avoids even the atomic pointer load in facade.
func BenchmarkNowInjected(b *testing.B) {
	c := Default()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Now()
	}
}
