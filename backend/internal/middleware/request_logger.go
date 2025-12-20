package middleware

import (
	"cso-erp/internal/logger"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestLoggerConfig holds configuration for request logger
type RequestLoggerConfig struct {
	// SkipPaths defines paths to skip logging
	SkipPaths []string
	// LogBody enables request/response body logging (use with caution)
	LogBody bool
}

// DefaultRequestLoggerConfig returns default config
func DefaultRequestLoggerConfig() RequestLoggerConfig {
	return RequestLoggerConfig{
		SkipPaths: []string{"/health", "/swagger"},
		LogBody:   false,
	}
}

// RequestLogger creates a middleware for structured request logging
func RequestLogger(cfg ...RequestLoggerConfig) fiber.Handler {
	var config RequestLoggerConfig
	if len(cfg) > 0 {
		config = cfg[0]
	} else {
		config = DefaultRequestLoggerConfig()
	}

	return func(c *fiber.Ctx) error {
		// Skip paths
		path := c.Path()
		for _, skip := range config.SkipPaths {
			if len(path) >= len(skip) && path[:len(skip)] == skip {
				return c.Next()
			}
		}

		// Generate request ID if not present
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()[:8]
		}

		// Set request ID in response header
		c.Set("X-Request-ID", requestID)

		// Store in locals for other handlers
		c.Locals("request_id", requestID)

		// Get user info from context (if authenticated)
		userID := ""
		if uid, ok := c.Locals("user_id").(uuid.UUID); ok {
			userID = uid.String()
		}

		tenantDB := ""
		if tdb, ok := c.Locals("tenant_db").(string); ok {
			tenantDB = tdb
		}

		// Create request logger
		reqLogger := logger.Default().
			WithRequestID(requestID).
			WithUserID(userID).
			WithTenant(tenantDB)

		// Store logger in context
		c.Locals("logger", reqLogger)

		// Start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(start)

		// Log request
		reqLogger.LogRequest(
			c.Method(),
			path,
			c.Response().StatusCode(),
			duration,
			GetClientIP(c),
			c.Get("User-Agent"),
		)

		return err
	}
}

// GetLogger returns the logger from fiber context
func GetLogger(c *fiber.Ctx) *logger.Logger {
	if log, ok := c.Locals("logger").(*logger.Logger); ok {
		return log
	}
	return logger.Default()
}

// GetRequestID returns the request ID from fiber context
func GetRequestID(c *fiber.Ctx) string {
	if reqID, ok := c.Locals("request_id").(string); ok {
		return reqID
	}
	return ""
}
