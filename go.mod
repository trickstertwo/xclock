module github.com/trickstertwo/xclock

go 1.25

require (
	github.com/trickstertwo/xclock/adapters/slogclock v0.0.0-20251005024325-d2c5180bff82
	github.com/trickstertwo/xclock/adapters/zapclock v0.0.0-20251005024325-d2c5180bff82
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.11.0 // indirect
