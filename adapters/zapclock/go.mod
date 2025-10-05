module github.com/trickstertwo/xclock/adapters/zapclock

go 1.25

require (
	github.com/trickstertwo/xclock v0.0.0
	go.uber.org/zap v1.27.0
)

require go.uber.org/multierr v1.10.0 // indirect

replace github.com/trickstertwo/xclock => ../..
