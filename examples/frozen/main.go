package main

import (
	"fmt"
	"time"

	"github.com/trickstertwo/xclock"
	"github.com/trickstertwo/xclock/adapter/frozen"
)

func main() {
	// Deterministic clock for tests/demos.
	frozen.Use(frozen.Config{
		Time: time.Date(2030, 1, 1, 0, 0, 0, 0, time.UTC),
	})

	fmt.Println("== frozen example ==")
	fmt.Println("Now()", xclock.Now().Format(time.RFC3339Nano))

	// Scheduling still uses real timers to avoid deadlocks if used in prod by mistake.
	fired := make(chan struct{}, 1)
	cancel := xclock.AfterFunc(15*time.Millisecond, func() {
		fired <- struct{}{}
	})
	defer cancel()

	select {
	case <-fired:
		fmt.Println("AfterFunc: fired")
	case <-time.After(200 * time.Millisecond):
		fmt.Println("AfterFunc: timed out")
	}
}
