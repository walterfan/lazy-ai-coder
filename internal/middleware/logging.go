package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/walterfan/lazy-ai-coder/internal/log"
)

// LoggingMiddleware provides structured logging with correlation IDs
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Get request ID
		requestID := GetRequestID(c)

		logger := log.GetLogger()

		// Log request start
		logger.With(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"query", c.Request.URL.RawQuery,
			"ip", c.ClientIP(),
			"user_agent", c.Request.UserAgent(),
		).Info("Request started")

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Log request completion
		status := c.Writer.Status()

		// Build base logger with fields
		loggerWithFields := logger.With(
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"duration", duration.String(),
			"duration_ms", duration.Milliseconds(),
			"size", c.Writer.Size(),
		)

		// Add errors if any
		if len(c.Errors) > 0 {
			loggerWithFields = loggerWithFields.With("errors", c.Errors.String())
		}

		// Choose log level based on status code
		switch {
		case status >= 500:
			loggerWithFields.Error("Request failed with server error")
		case status >= 400:
			loggerWithFields.Warn("Request failed with client error")
		case status >= 300:
			loggerWithFields.Info("Request redirected")
		default:
			loggerWithFields.Info("Request completed")
		}
	}
}
