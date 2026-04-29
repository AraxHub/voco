package main

import (
	"log/slog"
	"os"
	"voco/internal/app"
)

func main() {
	releaseMode := os.Getenv("VOCO_RELEASE_MODE")
	if releaseMode == "" {
		releaseMode = "prod"
	}

	cfg, err := app.MustLoadCfg(releaseMode)
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	a := app.NewApp(cfg)
	if err := a.Run(); err != nil {
		slog.Error("run failed", "error", err)
		os.Exit(1)
	}
}
