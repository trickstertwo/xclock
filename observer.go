package xclock

import (
	"sync"
	"time"
)

// Observer: subscribe to tick events from an observable ticker.

// TickObserver receives tick notifications.
type TickObserver interface {
	OnTick(t time.Time)
}

// ObservableTicker is a Ticker that supports observer subscriptions.
// It incurs a single goroutine to fan out ticks only when constructed.
type ObservableTicker interface {
	Ticker
	Subscribe(obs TickObserver) (unsubscribe func())
}

// NewObservableTicker returns an ObservableTicker backed by the provided clock.
// If c is nil, Default() is used.
func NewObservableTicker(d time.Duration, c Clock) ObservableTicker {
	if c == nil {
		c = Default()
	}
	return newObservableTicker(d, c)
}

type observableTicker struct {
	inner  Ticker
	mu     sync.RWMutex
	subs   map[TickObserver]struct{}
	quitCh chan struct{}
}

func newObservableTicker(d time.Duration, c Clock) *observableTicker {
	ot := &observableTicker{
		inner:  c.NewTicker(d),
		subs:   make(map[TickObserver]struct{}),
		quitCh: make(chan struct{}),
	}
	go ot.run()
	return ot
}

func (o *observableTicker) run() {
	for {
		select {
		case t := <-o.inner.C():
			o.mu.RLock()
			for s := range o.subs {
				// Best-effort, non-blocking via goroutine to isolate observers.
				// Keeps core clock path fast and dependable.
				s := s
				tt := t
				go s.OnTick(tt)
			}
			o.mu.RUnlock()
		case <-o.quitCh:
			return
		}
	}
}

func (o *observableTicker) C() <-chan time.Time { return o.inner.C() }

func (o *observableTicker) Stop() {
	close(o.quitCh)
	o.inner.Stop()
}

func (o *observableTicker) Reset(d time.Duration) { o.inner.Reset(d) }

func (o *observableTicker) Subscribe(obs TickObserver) (unsubscribe func()) {
	o.mu.Lock()
	o.subs[obs] = struct{}{}
	o.mu.Unlock()
	return func() {
		o.mu.Lock()
		delete(o.subs, obs)
		o.mu.Unlock()
	}
}
