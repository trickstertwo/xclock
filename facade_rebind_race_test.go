package xclock

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Ensures that rebinding the facade via SetDefault is race-free and does not panic
// under concurrent reads/writes. Run with: go test -race
func TestConcurrent_SetDefaultAndFacade_NoRace(t *testing.T) {
	orig := Default()
	t.Cleanup(func() { SetDefault(standardSystemClock); SetDefault(orig) })

	f1 := NewFrozen(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC))
	f2 := NewFrozen(time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC))

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup

	// Writers flip defaults
	for i := 0; i < 2; i++ {
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

	// Readers hammer facade
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for ctx.Err() == nil {
				_ = Now()
				_ = Since(Now())
				_ = AfterFunc(1*time.Nanosecond, func() {}) // exercise timer path
			}
		}()
	}

	wg.Wait()
}
