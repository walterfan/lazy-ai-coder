package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/walterfan/lazy-ai-coder/internal/metrics"
)

// MetricsMiddleware records HTTP request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath()

		// If path is empty (404), use the request path
		if path == "" {
			path = c.Request.URL.Path
		}

		metrics.RecordHTTPRequest(method, path, status, duration)
	}
}
