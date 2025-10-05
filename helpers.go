package xclock

import (
	"context"
	"time"
)

// Helper library: shared helpers for ergonomics. Kept separate for composability.

// SleepContext sleeps for d or until ctx is done. Returns context error if canceled.
func SleepContext(ctx context.Context, d time.Duration, c Clock) error {
	if c == nil {
		c = Default()
	}
	if d <= 0 {
		return ctx.Err()
	}
	timer := c.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C():
		return nil
	}
}

// Until periodically calls f every period until ctx is done.
// First call occurs after the first period elapses.
func Until(ctx context.Context, period time.Duration, f func(time.Time), c Clock) {
	if c == nil {
		c = Default()
	}
	t := c.NewTicker(period)
	defer t.Stop()
	for {
		select {
		case tt := <-t.C():
			f(tt)
		case <-ctx.Done():
			return
		}
	}
}
