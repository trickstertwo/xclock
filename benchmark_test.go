package xclock

import (
	"testing"
	"time"
)

func BenchmarkTimeNow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = time.Now()
	}
}

func BenchmarkNowFacadeSystem(b *testing.B) {
	// Ensure system fast-path
	orig := Default()
	defer SetDefault(orig)
	SetDefault(orig) // if system, binds stdlib

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Now()
	}
}

func BenchmarkNowInjectedClock(b *testing.B) {
	c := Default()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Now()
	}
}
