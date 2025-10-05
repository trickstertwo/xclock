package xclock

import (
	"testing"
	"time"
)

func TestBuilder_SystemDefault(t *testing.T) {
	t.Parallel()

	c := NewBuilder().Build()
	n := c.Now()
	// Expect near current time
	if n.Before(time.Now().Add(-5*time.Second)) || n.After(time.Now().Add(5*time.Second)) {
		t.Fatalf("builder system Now looks wrong: %v", n)
	}
}

func TestBuilder_Frozen(t *testing.T) {
	t.Parallel()

	ft := time.Date(2000, 1, 2, 3, 4, 5, 6, time.UTC)
	c := NewBuilder().Apply(WithStrategy(StrategyFrozen), WithFrozenTime(ft)).Build()
	if got := c.Now(); !got.Equal(ft) {
		t.Fatalf("frozen Now mismatch: got=%v want=%v", got, ft)
	}
}

func TestFactory_KnownProviders(t *testing.T) {
	t.Parallel()

	// system
	sys, err := NewFromFactory("system", Options{})
	if err != nil {
		t.Fatalf("system provider error: %v", err)
	}
	_ = sys.Now() // sanity

	// frozen
	ft := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	fr, err := NewFromFactory("frozen", Options{FrozenTime: ft})
	if err != nil {
		t.Fatalf("frozen provider error: %v", err)
	}
	if got := fr.Now(); !got.Equal(ft) {
		t.Fatalf("frozen provider Now mismatch: got=%v want=%v", got, ft)
	}
}

func TestFactory_Register(t *testing.T) {
	t.Parallel()

	name := "test-provider"
	Register(name, func(opts Options) (Clock, error) {
		return NewFrozen(opts.FrozenTime), nil
	})

	ft := time.Date(2010, 1, 1, 0, 0, 0, 0, time.UTC)
	c, err := NewFromFactory(name, Options{FrozenTime: ft})
	if err != nil {
		t.Fatalf("custom provider error: %v", err)
	}
	if got := c.Now(); !got.Equal(ft) {
		t.Fatalf("custom provider Now mismatch: got=%v want=%v", got, ft)
	}
}
