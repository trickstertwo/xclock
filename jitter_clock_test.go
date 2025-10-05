package xclock

import (
	"testing"
	"time"
)

func TestJitter_WithSeedDeterministic(t *testing.T) {
	t.Parallel()

	baseT := time.Date(2032, 1, 2, 3, 4, 5, 6, time.UTC)
	maxJ := 5 * time.Millisecond
	seed := uint64(123456789)

	clk1 := NewBuilder().Apply(
		WithStrategy(StrategyFrozen),
		WithFrozenTime(baseT),
		WithJitter(maxJ),
		WithJitterSeed(seed),
	).Build()

	clk2 := NewBuilder().Apply(
		WithStrategy(StrategyFrozen),
		WithFrozenTime(baseT),
		WithJitter(maxJ),
		WithJitterSeed(seed),
	).Build()

	const N = 16
	var seq1, seq2 [N]time.Duration
	for i := 0; i < N; i++ {
		seq1[i] = clk1.Now().Sub(baseT)
		seq2[i] = clk2.Now().Sub(baseT)

		// Bound check
		if seq1[i] < -maxJ || seq1[i] > maxJ {
			t.Fatalf("seq1[%d] out of bounds: %v (max %v)", i, seq1[i], maxJ)
		}
		if seq2[i] < -maxJ || seq2[i] > maxJ {
			t.Fatalf("seq2[%d] out of bounds: %v (max %v)", i, seq2[i], maxJ)
		}
		// Deterministic equality
		if seq1[i] != seq2[i] {
			t.Fatalf("deterministic sequences differ at %d: %v vs %v", i, seq1[i], seq2[i])
		}
	}
}

func TestJitter_ZeroOrNegativeDisables(t *testing.T) {
	t.Parallel()

	baseT := time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC)
	base := NewBuilder().Apply(WithStrategy(StrategyFrozen), WithFrozenTime(baseT)).Build()

	if got := NewJitter(base, 0).Now(); !got.Equal(baseT) {
		t.Fatalf("jitter(0) should return base time: got %v want %v", got, baseT)
	}
	if got := newJitterWithSeed(base, -1, 42).Now(); !got.Equal(baseT) { // negative treated as disabled
		t.Fatalf("jitter(-1) should return base time: got %v want %v", got, baseT)
	}
}
