// Package xclock provides a dependable, minimal, and stable clock facade
// for Go applications. It exposes a small Strategy interface (Clock) plus
// branchless facade functions (Now/Sleep/After/AfterFunc/NewTimer/NewTicker)
// and keeps the platform layer wrapped.
//
// Architecture (no circular deps)
//   - Core (this package): interfaces, singleton (Default/SetDefault), facade,
//     helpers, and observer. It does NOT know about concrete clocks.
//   - Adapters (submodules under adapter/...): implement xclock.Clock and expose
//     a Use(Config) function to SetDefault(...) explicitly, mirroring xlog.
//
// Design patterns
// 1) Singleton  – Default()/SetDefault().
// 2) Builder    – adapter/compose.Use(...) composes base + layers.
// 3) Factory    – adapter packages expose New(...) / Use(...).
// 4) Facade     – xclock.Now()/Sleep()/After()... are branchless to the current Clock.
// 5) Adapter    – adapter/* translate concrete time providers to xclock.Clock.
// 6) Strategy   – Clock is the swappable algorithm for time source.
// 7) Observer   – ObservableTicker to fan-out tick notifications.
//
// Usage
//
//	// 1) Choose an adapter explicitly (no env magic, no blank-imports).
//	//    Example: compose system base + offset + jitter.
//	compose.Use(compose.Config{
//	    Strategy:   compose.StrategySystem,
//	    Offset:     50 * time.Millisecond,
//	    Jitter:     2 * time.Millisecond,
//	    JitterSeed: 12345,
//	})
//
//	// 2) Call the facade anywhere in your code.
//	t := xclock.Now()
//	xclock.Sleep(10 * time.Millisecond)
//	ch := xclock.After(5 * time.Millisecond)
package xclock
