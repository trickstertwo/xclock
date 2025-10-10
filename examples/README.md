# xclock examples

These examples show the recommended, explicit `adapter.Use(...)` pattern (mirrors `xlog`) with a single call that selects the default clock for your process. No envs, no blank-imports.

- `compose/` – Optimal, single-Use composition for most apps (choose a base strategy, optional offset/jitter).
- `frozen/` – Deterministic time source for tests; scheduling still uses real timers.
- `calibrated/` – Dynamic offset learned from an authority (e.g., NTP/secure time/GPS).
- `system/` – Explicit system clock.

## Run

From the repo root (workspace has `go.work` including `examples`):

```bash
go run ./examples/compose
go run ./examples/frozen
go run ./examples/calibrated
go run ./examples/system
```

Each example preserves and restores the previous default to remain isolated.