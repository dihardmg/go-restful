package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	// TraceIDKey is the key used to store trace ID in Gin context
	TraceIDKey = "trace_id"
)

// Logger is a middleware that logs requests with trace ID
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Generate trace ID for this request
		traceID := uuid.New().String()

		// Store trace ID in context for use in handlers
		c.Set(TraceIDKey, traceID)

		// Add trace ID to response header for client tracking
		c.Header("X-Trace-ID", traceID)

		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log request details with trace ID
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path

		// Format duration consistently (always show in ms with 2 decimals)
		var durationStr string
		if duration.Milliseconds() > 0 {
			durationStr = fmt.Sprintf("%dms", duration.Milliseconds())
		} else {
			durationStr = fmt.Sprintf("%.2fµs", float64(duration.Microseconds()))
		}

		// Determine log level based on status code
		level := "INFO"
		if statusCode >= 500 {
			level = "ERROR"
		} else if statusCode >= 400 {
			level = "WARN"
		}

		log.Printf("[%s] [%s] %s | %3d | %8s | %15s | %-7s %s",
			traceID,
			level,
			time.Now().Format("2006/01/02 15:04:05"),
			statusCode,
			durationStr,
			clientIP,
			method,
			path,
		)

		// Log errors if any
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				log.Printf("[%s] [ERROR] %s", traceID, e.Error())
			}
		}
	}
}
