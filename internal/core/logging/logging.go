package logging

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/lith-project/lith/internal/core/config"
)

// New builds a slog.Logger writing to w, configured by cfg.
func New(w io.Writer, cfg config.Log) (*slog.Logger, error) {
	// Parse level string
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		return nil, fmt.Errorf("logging: unknown level %q", cfg.Level)
	}

	// Create handler options
	opts := slog.HandlerOptions{
		Level: level,
	}

	// Create handler based on format
	var handler slog.Handler
	switch cfg.Format {
	case "text":
		handler = slog.NewTextHandler(w, &opts)
	case "json":
		handler = slog.NewJSONHandler(w, &opts)
	default:
		return nil, fmt.Errorf("logging: unknown format %q", cfg.Format)
	}

	return slog.New(handler), nil
}
