package main

import (
	"context"
	"fmt"
	"time"

	"github.com/trickstertwo/xclock"
)

type printer struct{}

func (p printer) OnTick(t time.Time) { fmt.Println("tick at", t) }

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	ot := xclock.NewObservableTicker(250*time.Millisecond, nil)
	defer ot.Stop()
	unsub := ot.Subscribe(printer{})
	defer unsub()

	// Also consume the ticker channel if desired
	go func() {
		for range ot.C() {
			// no-op
		}
	}()

	<-ctx.Done()
}
