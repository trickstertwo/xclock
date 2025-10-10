package main

import (
	"fmt"
	"time"

	"github.com/trickstertwo/xclock"
)

// This example compares stdlib time.Sleep vs xclock.Sleep using their
// respective time sources (time.Now vs xclock.Now).
func main() {
	d := 10 * time.Millisecond
	n := 100000

	fmt.Printf("Comparing sleep for %s over %d iterations\n", d, n)

	timeAvg := bench("time.Sleep + time.Now", time.Sleep, time.Now, time.Since, d, n)
	xclkAvg := bench("xclock.Sleep + xclock.Now", xclock.Sleep, xclock.Now, xclock.Since, d, n)

	fmt.Println()
	fmt.Printf("Single-shot with xclock:\n")
	start := xclock.Now()
	xclock.Sleep(d)
	elapsed := xclock.Since(start)
	fmt.Printf("elapsed: %s (target %s)\n", elapsed, d)

	fmt.Println()
	fmt.Println("Summary:")
	fmt.Printf("time avg:   %s\n", timeAvg)
	fmt.Printf("xclock avg: %s\n", xclkAvg)
}

func bench(
	name string,
	sleep func(time.Duration),
	now func() time.Time,
	since func(time.Time) time.Duration,
	d time.Duration,
	iters int,
) time.Duration {
	var total time.Duration
	for i := 0; i < iters; i++ {
		start := now()
		sleep(d)
		total += since(start)
	}
	avg := total / time.Duration(iters)
	fmt.Printf("%s => target=%s avg_elapsed=%s\n", name, d, avg)
	return avg
}
