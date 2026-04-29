package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Format  string `envconfig:"FORMAT" default:"text"` // text|json
	Level   string `envconfig:"LEVEL" default:"info"`  // debug|info|warn|error
	Service string `envconfig:"SERVICE" default:"voco-backend"`
	Env     string `envconfig:"ENV" default:"prod"`
}

func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)

	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler
	switch strings.ToLower(cfg.Format) {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, opts)
	default:
		h = slog.NewTextHandler(os.Stdout, opts)
	}

	log := slog.New(h).With(
		"service", cfg.Service,
		"env", cfg.Env,
	)
	return log
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "", "info":
		return slog.LevelInfo
	default:
		// Фолбэк на info, чтобы не падать от опечатки.
		return slog.LevelInfo
	}
}
