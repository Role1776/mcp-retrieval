package logger

import (
	"log/slog"
	"os"
)

const (
	modeLocal = "local"
	modeProd  = "prod"
)

type Config struct {
	Mode string `env:"LOG_MODE" envDefault:"local"`
}

func SetupLogger(cfg *Config) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{}

	switch cfg.Mode {
	case modeLocal:
		opts.Level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stderr, opts)
	case modeProd:
		opts.Level = slog.LevelInfo
		handler = slog.NewJSONHandler(os.Stderr, opts)
	default:
		opts.Level = slog.LevelDebug
		handler = slog.NewJSONHandler(os.Stderr, opts)
	}

	return slog.New(handler)
}
