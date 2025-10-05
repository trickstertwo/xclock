package main

import (
	"log/slog"
	"os"

	"github.com/trickstertwo/xclock"
	"github.com/trickstertwo/xclock/adapters/slogclock"
)

func main() {
	clk := xclock.Default()
	h := slogclock.WithClock(slog.NewJSONHandler(os.Stdout, nil), clk)
	log := slog.New(h)

	log.Info("hello with xclock time")
}
