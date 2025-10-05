// Package xclock provides a high-performance, dependency-injected clock framework
// designed for modular Go systems.
//
// Goals
// - Dependability: Zero-GC hot path, race-free, testable, production-safe.
// - Extendability: Strategy-driven clock sources; plugin registry; adapters.
// - Team scalability: Small surface, clear responsibilities, black-box modules.
// - Development velocity: Simple Facade API, functional options, sane defaults.
// - Risk reduction: Platform layer wrapped; swapping strategies/hardware safe.
//
// Design patterns used
// 1. Singleton: Default() global clock instance via atomic swap.
// 2. Builder: NewBuilder() with functional options → Build().
// 3. Factory: Registry NewFromFactory(name, ...Option) to construct clocks.
// 4. Facade: Top-level functions Now/Sleep/After delegate to Default().
// 5. Adapter: Subpackages in adapters/ integrate with logging ecosystems.
// 6. Strategy: Multiple clock implementations (system, frozen).
// 7. Observer: ObservableTicker notifies subscribers on ticks.
//
// Performance notes
// - Global clock stored in atomic.Value for lock-free reads.
// - System clock uses stdlib time primitives directly.
// - No generics or reflection in hot paths.
// - ObservableTicker is opt-in; no background goroutines unless used.
//
// Thread safety
// - All exported implementations are safe for concurrent use.
//
// Usage
//
//	// Use defaults
//	t := xclock.Now()
//
//	// Swap default clock (e.g., in tests)
//	xclock.SetDefault(xclock.NewFrozen(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)))
//
//	// Use with zap
//	logger := zap.New(zapcore.NewCore(...), zap.WithClock(adapterszap.New(xclock.Default())))
package xclock
