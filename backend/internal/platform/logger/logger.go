// Package logger builds the application-wide structured logger (log/slog).
package logger

import (
	"log/slog"
	"os"
	"strings"
)

// New returns a JSON structured logger at the requested level.
// Unknown levels fall back to info.
func New(level string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLevel(level)})
	return slog.New(handler)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
