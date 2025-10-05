package xclock

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type testObserver struct {
	count atomic.Int32
	wg    *sync.WaitGroup
}

func (o *testObserver) OnTick(_ time.Time) {
	o.count.Add(1)
	if o.wg != nil {
		o.wg.Done()
	}
}

func TestObservableTicker_SubscribeAndStop(t *testing.T) {
	t.Parallel()

	ot := NewObservableTicker(20*time.Millisecond, nil)
	defer ot.Stop()

	var wg sync.WaitGroup
	wg.Add(3)
	obs := &testObserver{wg: &wg}
	unsub := ot.Subscribe(obs)
	defer unsub()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(400 * time.Millisecond):
		t.Fatal("did not receive expected ticks in time")
	}

	// Unsubscribe and ensure no rapid additional increments in a short window.
	unsub()
	before := obs.count.Load()
	time.Sleep(80 * time.Millisecond)
	after := obs.count.Load()
	if delta := after - before; delta > 3 { // allow a small race window tolerance
		t.Fatalf("observer received too many ticks after unsubscribe: delta=%d", delta)
	}
}
