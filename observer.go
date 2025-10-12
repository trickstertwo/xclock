package xclock

import (
	"context"
	"sync"
	"time"
)

// TickObserver defines the observer interface for tick notifications.
// Subscribers implement this to receive non-blocking tick events.
type TickObserver interface {
	OnTick(time.Time)
}

// ObservableTicker fans out ticker notifications to multiple observers.
// It implements the Observer pattern for tick events, with opt-in background goroutine.
// Start begins fan-out; Stop ends it. Observers are added/removed dynamically.
type ObservableTicker struct {
	ticker    Ticker
	observers map[TickObserver]bool
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	done      chan struct{}
}

// NewObservableTicker creates a new ObservableTicker with the given interval.
// No background goroutine starts until Start() is called.
func NewObservableTicker(d time.Duration) *ObservableTicker {
	return &ObservableTicker{
		ticker:    NewTicker(d),
		observers: make(map[TickObserver]bool),
		done:      make(chan struct{}),
	}
}

// AddObserver adds an observer to receive tick notifications.
// Thread-safe; can be called concurrently.
func (ot *ObservableTicker) AddObserver(obs TickObserver) {
	ot.mu.Lock()
	defer ot.mu.Unlock()
	ot.observers[obs] = true
}

// RemoveObserver removes an observer.
// Thread-safe; can be called concurrently.
func (ot *ObservableTicker) RemoveObserver(obs TickObserver) {
	ot.mu.Lock()
	defer ot.mu.Unlock()
	delete(ot.observers, obs)
}

// Start begins the background fan-out goroutine.
// Ticks are forwarded to all observers via OnTick.
// Call Stop to end.
func (ot *ObservableTicker) Start() {
	ot.mu.Lock()
	if ot.cancel != nil {
		ot.mu.Unlock()
		return // already started
	}
	ot.ctx, ot.cancel = context.WithCancel(context.Background())
	ot.mu.Unlock()

	go func() {
		defer close(ot.done)
		for {
			select {
			case t := <-ot.ticker.C():
				ot.mu.RLock()
				for obs := range ot.observers {
					go obs.OnTick(t) // non-blocking fan-out
				}
				ot.mu.RUnlock()
			case <-ot.ctx.Done():
				return
			}
		}
	}()
}

// Stop ends the ticker and fan-out goroutine.
// Safe to call multiple times.
func (ot *ObservableTicker) Stop() {
	ot.mu.Lock()
	if ot.cancel != nil {
		ot.cancel()
		ot.cancel = nil
	}
	ot.mu.Unlock()
	ot.ticker.Stop()
	<-ot.done // wait for goroutine to exit
}
