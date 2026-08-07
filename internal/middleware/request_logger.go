package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger returns a Gin middleware that logs each HTTP request
// using the global slog logger with structured fields.
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Choose log level based on status code
		switch {
		case status >= 500:
			slog.Error("http request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"client_ip", c.ClientIP(),
			)
		case status >= 400:
			slog.Warn("http request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"client_ip", c.ClientIP(),
			)
		default:
			slog.Info("http request",
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
				"client_ip", c.ClientIP(),
			)
		}
	}
}
