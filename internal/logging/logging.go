// Package logging provides centralized logging for the application.
package logging

import (
	"log/slog"
	"os"
)

// NewJSONLogger returns a slog.Logger configured for JSON output.
func NewJSONLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
