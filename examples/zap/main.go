package main

import (
	"os"

	"github.com/trickstertwo/xclock"
	"github.com/trickstertwo/xclock/adapters/zapclock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	clock := xclock.Default()

	encCfg := zap.NewProductionEncoderConfig()
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encCfg),
		zapcore.AddSync(os.Stdout),
		zap.InfoLevel,
	)

	log := zap.New(core, zap.WithClock(zapclock.New(clock)))
	defer func(log *zap.Logger) {
		if err := log.Sync(); err != nil {
			panic(err)
		}
	}(log)

	log.Info("hello with xclock time")
}
