package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// RequestIDKey is the key used to store request ID in context
	RequestIDKey = "request_id"

	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"

	// CorrelationIDHeader is the header name for correlation ID (from client)
	CorrelationIDHeader = "X-Correlation-ID"
)

// RequestIDMiddleware generates or extracts request ID for tracing
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get correlation ID from header first (for distributed tracing)
		requestID := c.GetHeader(CorrelationIDHeader)

		// If no correlation ID from client, check for X-Request-ID
		if requestID == "" {
			requestID = c.GetHeader(RequestIDHeader)
		}

		// If still no ID, generate a new one
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context for downstream use
		c.Set(RequestIDKey, requestID)

		// Add to response headers for client-side tracing
		c.Header(RequestIDHeader, requestID)
		c.Header(CorrelationIDHeader, requestID)

		c.Next()
	}
}

// GetRequestID extracts request ID from gin context
func GetRequestID(c *gin.Context) string {
	if requestID, exists := c.Get(RequestIDKey); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
