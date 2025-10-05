# xclock

High-performance, modular clock framework for Go. Branchless, zero-alloc hot paths with pluggable strategies and lean logging adapters.

## Why xclock

- Dependability
    - Branchless facade bound via function pointers; zero allocations on hot paths.
    - Lock-free reads; atomic pointer swap on SetDefault; race-safe.
    - Standard library-backed system strategy is the default fast-path.

- Extendability and modularity
    - Minimal Clock interface (Strategy pattern). System, Frozen, Offset, Jitter, and Calibrated strategies provided.
    - Builder and Factory for composition and plug-in providers.
    - Adapters split into separate submodules (zap, slog) to minimize dependency blast radius.

- Team scalability and velocity
    - Tiny, stable API; DI-friendly; “one module per person”.
    - Helpers and ObservableTicker centralize common timing patterns.

- Risk reduction
    - Platform/time source wrapped; swap strategies without touching call sites.
    - Adapters isolated from core; logging libs can change without affecting core.

## Design patterns used

1. Singleton – process-wide default clock via Default and SetDefault.
2. Builder – NewBuilder().Apply(...).Build() composes strategies step by step.
3. Factory – Register and NewFromFactory for provider-based construction.
4. Facade – top-level functions Now, Sleep, After, NewTimer, NewTicker, AfterFunc.
5. Adapter – submodules for zap and slog integrations.
6. Strategy – interchangeable Clock implementations (system, frozen, offset, jitter, calibrated).
7. Observer – ObservableTicker + TickObserver for non-blocking tick subscribers.

## Architecture highlights

- Facade uses function pointers swapped in SetDefault:
    - System default binds directly to time.Now/time.Sleep/etc. (stdlib fast-path).
    - Custom clocks bind to injected methods.
    - No branches on the hot path; one atomic pointer load + direct call.
- atomic.Value stores a stable wrapper type to avoid “inconsistently typed value” panics.
- No background goroutines unless you opt-in (e.g., ObservableTicker, Calibrated auto-sync).

## Modules

- Core (this module)
    - Clock interface with strategies: System, Frozen, Offset, Jitter (deterministic seed support), Calibrated (dynamic offset).
    - Facade, Builder, Factory, ObservableTicker (Observer), Helpers (SleepContext, Until).
- Adapters (independent modules)
    - github.com/trickstertwo/xclock/adapters/zapclock – implements zapcore.Clock
    - github.com/trickstertwo/xclock/adapters/slogclock – wraps slog.Handler

## Install

- Core:
  go get github.com/trickstertwo/xclock

- Zap adapter:
  go get github.com/trickstertwo/xclock/adapters/zapclock

- Slog adapter:
  go get github.com/trickstertwo/xclock/adapters/slogclock

## Quick start

```go
// Facade (system fast-path is branchless and zero-alloc)
t0 := xclock.Now()
xclock.Sleep(5*time.Millisecond)
elapsed := xclock.Since(t0)

// Dependency injection (capture Clock once for hot loops)
clk := xclock.Default()
timer := clk.NewTimer(10 * time.Millisecond)
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

## Pluggable strategies (future-proof)

Build and compose without changing call sites (e.g., your logger):

```go
// Compose offset (+50ms) and jitter (±5ms) deterministically for tests/simulations:
clk := xclock.NewBuilder().Apply(
  xclock.WithOffset(50*time.Millisecond),
  xclock.WithJitter(5*time.Millisecond),
  xclock.WithJitterSeed(123456789), // deterministic sequences
).Build()
xclock.SetDefault(clk)

// Calibrated time (e.g., from NTP/secure source), dynamic offset:
cal := xclock.NewCalibrated(nil) // base = system
_ = cal.SyncOnce(context.Background(), func(ctx context.Context) (time.Time, error) {
  // Fetch authoritative time...
  return time.Now().Add(120 * time.Millisecond), nil
})
xclock.SetDefault(cal)
```

Available strategies:
- System: stdlib-backed fast-path, zero overhead.
- Frozen: deterministic timestamps for tests.
- Offset: fixed shift of reported wall time; scheduling delegates to base.
- Jitter: symmetric jitter in [-max,+max]; deterministic via WithJitterSeed.
- Calibrated: dynamic offset learned from an external authority; optional periodic auto-sync.

## Examples

- examples/basic – facade and DI usage
- examples/zap – zap integration
- examples/slog – slog integration
- examples/observer – observable ticker with subscribers
- examples/strategies – composing offset/jitter/calibrated without changing call sites

## Performance

- Facade: one atomic pointer load + direct function call; stdlib fast-path for system clock.
- Zero allocations on hot paths (tests assert this).
- For ultra-hot loops, inject and capture Clock to avoid even the atomic load.
- Optional features (ObservableTicker, Calibrated auto-sync) are opt-in.

Benchmarks (representative; run locally):
- BenchmarkTimeNow
- BenchmarkNowFacade_System
- BenchmarkNowInjected

## Testing

- go test ./...
- go test -race ./... (recommended)

Coverage includes:
- Facade fast-path, rebinding on SetDefault, and concurrency safety.
- Strategies: System, Frozen, Offset, Jitter (deterministic), Calibrated (delta application).
- Timers, Tickers, AfterFunc semantics.
- Helpers (SleepContext, Until).
- ObservableTicker subscribe/unsubscribe behavior.
- Adapters (zap/slog) stamping with injected Clock.

## Versioning and compatibility

- Semantic versioning; stable Clock interface.
- Adapters version independently; core has no logging deps.
- Prefer tagged versions for adapters (avoid local replaces for portability).

## Roadmap

- Additional providers (secure/monotonic-only, NTP integration).
- Optional metrics hooks via Observer pattern.
- CI matrix across modules (core and adapters), race and allocation checks.

---
By standardizing on xclock, you get consistent timestamps across frameworks, deterministic tests, and a pluggable time policy—all without sacrificing performance or coupling your code to a specific platform or logging library.