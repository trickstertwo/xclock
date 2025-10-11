# xclock

High-performance, modular clock framework for Go. Branchless facade, stable interfaces, and pluggable time sources via adapters. Designed for dependability, extendability, team scalability, and development velocity with reduced risk.

## Why xclock

- Dependability
    - Branchless facade: function pointers bound atomically; one atomic load + one call on hot paths.
    - Lock-free reads; race-safe SetDefault with atomic swap and immediate facade rebind.
    - Standard library-backed system clock is the fast default; no allocations on facade calls.

- Extendability and modularity
    - Minimal Clock interface (Strategy pattern). Time sources provided as adapters: frozen, offset, jitter, calibrated; composition via a compose adapter.
    - Platform layer wrapped; adapters are black boxes that import xclock (no circular dependencies).
    - Helpers (SleepContext, Until) and ObservableTicker (Observer) consolidate common timing patterns.

- Team scalability and velocity
    - Tiny, stable API surface; DI-friendly; “one module per person” with adapters.
    - Explicit Use(...) configuration mirrors the xlog style; no env/blank-import magic.

- Risk reduction
    - Encapsulate platform/time-source changes inside adapters; swap at startup without touching call sites.
    - Clear boundaries and contracts via interfaces and adapters.

## Design patterns used

1. Singleton – process-wide default via Default and SetDefault with atomic swap.
2. Builder – adapter/compose.Use(...) composes base + layers step by step.
3. Factory – adapter packages expose New(...) and Use(...) to construct clocks.
4. Facade – top-level functions Now, Sleep, After, NewTimer, NewTicker, AfterFunc.
5. Adapter – time-source implementations live under adapter/* and translate external APIs to Clock.
6. Strategy – interchangeable Clock implementations (system, frozen, offset, jitter, calibrated).
7. Observer – ObservableTicker + TickObserver for non-blocking tick subscribers.

## Architecture

- Core (this package):
    - Stable interfaces: Clock, Timer, Ticker.
    - Facade functions bound via atomic function pointers.
    - Process-wide Default/SetDefault (Singleton).
    - System clock (stdlib) as dependable baseline fast-path.
    - Helpers and ObservableTicker.
    - No knowledge of adapters.

- Adapters (subpackages; black boxes):
    - adapter/frozen – deterministic Now/Since; scheduling delegates to real timers.
    - adapter/offset – fixed offset overlay on base.
    - adapter/jitter – symmetric jitter overlay on base (deterministic with seed).
    - adapter/calibrated – dynamic offset with SyncOnce/StartAutoSync.
    - adapter/compose – builder-style composition and a Use(...) that sets Default().
    - All adapters import xclock; xclock does not import adapters.

No background goroutines unless you opt-in (e.g., ObservableTicker fan-out, calibrated auto-sync).

## Install

- Core:
    - `go get github.com/trickstertwo/xclock`
- Adapters (import as needed; go will fetch automatically):
    - `github.com/trickstertwo/xclock/adapter/frozen`
    - `github.com/trickstertwo/xclock/adapter/offset`
    - `github.com/trickstertwo/xclock/adapter/jitter`
    - `github.com/trickstertwo/xclock/adapter/calibrated`
    - `github.com/trickstertwo/xclock/adapter/compose`

## Quick start

```go
// System fast-path (stdlib), branchless and zero-alloc:
t0 := xclock.Now()
xclock.Sleep(5 * time.Millisecond)
elapsed := xclock.Since(t0)

// Dependency injection (capture Clock once for hot loops):
clk := xclock.Default()
tm := clk.NewTimer(10 * time.Millisecond)
<-tm.C()
tm.Stop()
```

Compose a default clock (mirrors xlog.Use(...) style) using adapter/compose:

```go
import (
  "time"
  "github.com/trickstertwo/xclock"
  "github.com/trickstertwo/xclock/adapter/compose"
)

func main() {
  compose.Use(compose.Config{
    Strategy:   compose.StrategySystem, // or StrategyFrozen with FrozenTime
    Offset:     50 * time.Millisecond,
    Jitter:     2 * time.Millisecond,
    JitterSeed: 42, // optional, for deterministic jitter
  })

  t := xclock.Now()
  _ = t
}
```

## Examples

Run from repo root (go.work includes examples):

```bash
go run ./examples/frozen
go run ./examples/sleep_compare
```

Frozen (deterministic timestamps; scheduling remains real-time by design):

```go
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

  // Scheduling still uses real timers to avoid deadlocks.
  fired := make(chan struct{}, 1)
  cancel := xclock.AfterFunc(15*time.Millisecond, func() { fired <- struct{}{} })
  defer cancel()

  select {
  case <-fired:
    fmt.Println("AfterFunc: fired")
  case <-time.After(200 * time.Millisecond):
    fmt.Println("AfterFunc: timed out")
  }
}
```

Quick comparison (sleep and time source):

```go
package main

import (
  "fmt"
  "time"

  "github.com/trickstertwo/xclock"
)

// Compare stdlib time.Sleep/time.Now vs xclock facade.
func main() {
  d := 10 * time.Millisecond
  n := 1000

  fmt.Printf("Comparing sleep for %s over %d iterations\n", d, n)

  timeAvg := bench("time.Sleep + time.Now", time.Sleep, time.Now, time.Since, d, n)
  xclkAvg := bench("xclock.Sleep + xclock.Now", xclock.Sleep, xclock.Now, xclock.Since, d, n)

  fmt.Println()
  fmt.Println("Single-shot with xclock:")
  start := xclock.Now()
  xclock.Sleep(d)
  fmt.Printf("elapsed: %s (target %s)\n", xclock.Since(start), d)

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
```

## Calibrated time (dynamic offset)

```go
import (
  "context"
  "time"

  "github.com/trickstertwo/xclock"
  "github.com/trickstertwo/xclock/adapter/calibrated"
)

func main() {
  c := calibrated.New(nil) // base = xclock.Default() (system unless changed)
  _ = c.SyncOnce(context.Background(), func(ctx context.Context) (time.Time, error) {
    // pretend authoritative time is ahead by 120ms
    return time.Now().Add(120 * time.Millisecond), nil
  })
  xclock.SetDefault(c)
}
```

## Performance

- Facade: one atomic pointer load + direct function call.
- Zero allocations on hot paths.
- At very small sleeps (1ms), OS timer/scheduler jitter dominates; facade overhead is in nanoseconds.
- For ultra-hot loops, capture a Clock instance once and call methods directly to avoid the atomic load.

## Versioning

- Follow SemVer with git tags (source of truth).
- For runtime metadata, `xclock.Version()` returns module version when available via build info; otherwise "devel".

## Testing

- `go test ./...` and `go test -race ./...`
- Coverage includes facade rebinding, concurrency safety, timers/tickers, helpers, observer, and adapter behaviors.

---

By standardizing on xclock, you get consistent, deterministic, and swappable time behavior across your system with a tiny, dependable core and explicit, low-risk adapters.