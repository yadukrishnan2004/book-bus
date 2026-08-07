package logger

import (
	"log/slog"
	"os"
)

// Init sets up the global slog logger.
// In development (GIN_MODE != "release") it uses a human-readable text format.
// In production (GIN_MODE=release) it outputs structured JSON.
func Init() {
	var handler slog.Handler

	ginMode := os.Getenv("GIN_MODE")

	if ginMode == "release" {
		// Production: structured JSON — easy to pipe into Datadog, Loki, etc.
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Development: human-readable text with full debug output
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	slog.SetDefault(slog.New(handler))
}
