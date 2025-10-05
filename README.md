# xclock

High-performance, modular clock framework for Go. Zero-alloc, branchless facade with pluggable strategies and logging adapters.

## Why xclock

- Dependability
    - Branchless facade bound via function pointers; zero allocs on hot paths.
    - Lock-free reads; atomic pointer swap on SetDefault.
    - Standard library-backed system strategy by default.

- Extendability and modularity
    - Clock interface with Strategy pattern (system, frozen; easy to add more).
    - Factory/registry to plug custom providers.
    - Adapters split into separate submodules (zap, slog) to minimize deps.

- Team scalability and velocity
    - Tiny stable API; black-box modules; DI-friendly.
    - Helpers and observable ticker (Observer pattern) consolidate common patterns.

- Risk reduction
    - Platform/time source wrapped; safe to swap strategies.
    - Adapters isolated; core stays lean even if logging deps churn.

## Design patterns used

1. Singleton: process-wide default clock (`Default`, `SetDefault`).
2. Builder: `NewBuilder().Apply(...).Build()` for constructing strategies.
3. Factory: provider registry (`Register`, `NewFromFactory`).
4. Facade: top-level functions (`Now`, `Sleep`, `After`, `NewTimer`, `NewTicker`, `AfterFunc`).
5. Adapter: zap and slog integrations in separate submodules.
6. Strategy: interchangeable clock implementations (system, frozen).
7. Observer: `NewObservableTicker` with `TickObserver`.

## Architecture highlights

- Facade is bound to function pointers swapped in `SetDefault`. With the system clock it calls the stdlib directly (`time.Now`, etc.). With custom clocks it calls the injected methods. No branches on the hot path.
- `atomic.Value` stores a stable wrapper type to avoid inconsistent type panics.
- No background goroutines unless you opt-in (e.g., `ObservableTicker`).

## Modules

- Core (this module)
    - `xclock.Clock` interface, system and frozen strategies
    - Facade, builder, factory, observer ticker, helpers
- Adapters (independent modules)
    - `github.com/trickstertwo/xclock/adapters/zapclock` (zapcore.Clock)
    - `github.com/trickstertwo/xclock/adapters/slogclock` (slog Handler wrapper)

## Install

- Core:
  go get github.com/trickstertwo/xclock

- Zap adapter:
  go get github.com/trickstertwo/xclock/adapters/zapclock

- Slog adapter:
  go get github.com/trickstertwo/xclock/adapters/slogclock

## Quick start

```go
// Facade (fast-path on system)
t := xclock.Now()
xclock.Sleep(5*time.Millisecond)
elapsed := xclock.Since(t)

// Dependency injection
clk := xclock.Default()
timer := clk.NewTimer(10*time.Millisecond)
<-timer.C()

// Testing (deterministic time)
xclock.SetDefault(xclock.NewFrozen(time.Date(2025,1,1,0,0,0,0,time.UTC)))
defer xclock.SetDefault(xclock.Default()) // restore in your test harness
```

Adapters:

```go
// zap
log := zap.New(core, zap.WithClock(zapclock.New(xclock.Default())))

// slog
h := slogclock.WithClock(slog.NewJSONHandler(os.Stdout, nil), xclock.Default())
log := slog.New(h)
```

## Examples

- examples/basic: facade and DI usage
- examples/zap: zap integration
- examples/slog: slog integration
- examples/observer: observable ticker with subscribers

## Performance

- Facade calls: 1 atomic pointer load + direct function call; stdlib fast-path for system clock.
- No allocations on hot paths.
- Prefer capturing `xclock.Clock` in hot loops to avoid even the atomic load.

Benchmarks (representative; run with your environment):

- `BenchmarkTimeNow`
- `BenchmarkNowFacadeSystem`
- `BenchmarkNowInjectedClock`

## Testing

- Unit tests cover:
    - Facade fast-path and rebinding on `SetDefault`
    - Frozen/system strategies, timers, tickers, `AfterFunc`
    - Concurrency safety: concurrent `SetDefault` and facade calls
    - Helpers and observable ticker
    - Adapters (zap/slog) time stamping

Run:

- go test ./...
- go test -race ./... (recommended)

## Versioning and compatibility

- Semantic versioning; stable `Clock` interface.
- Submodules version independently to limit dependency blast radius.

## Roadmap

- Additional strategies (e.g., secure/monotonic-only/offset/jitter)
- Optional metrics hooks via Observer pattern
- CI matrix for modules and adapters