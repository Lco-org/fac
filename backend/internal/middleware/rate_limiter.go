package middleware

import (
	"cso-erp/internal/config"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type rateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	max      int
	window   time.Duration
}

func newRateLimiter(max int, window time.Duration) *rateLimiter {
	rl := &rateLimiter{
		requests: make(map[string][]time.Time),
		max:      max,
		window:   window,
	}

	// Cleanup old entries every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			rl.cleanup()
		}
	}()

	return rl
}

func (rl *rateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for ip, times := range rl.requests {
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(rl.requests, ip)
		} else {
			rl.requests[ip] = valid
		}
	}
}

func (rl *rateLimiter) allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()

	// Get request times for this IP
	times := rl.requests[ip]

	// Remove old requests outside the window
	var valid []time.Time
	for _, t := range times {
		if now.Sub(t) < rl.window {
			valid = append(valid, t)
		}
	}

	// Check if limit exceeded
	if len(valid) >= rl.max {
		return false
	}

	// Add current request
	valid = append(valid, now)
	rl.requests[ip] = valid

	return true
}

// RateLimitMiddleware limits requests per IP
func RateLimitMiddleware(cfg *config.SecurityConfig) fiber.Handler {
	limiter := newRateLimiter(cfg.RateLimitMax, cfg.RateLimitWindow)

	return func(c *fiber.Ctx) error {
		ip := c.IP()

		if !limiter.allow(ip) {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded. Please try again later.",
			})
		}

		return c.Next()
	}
}
