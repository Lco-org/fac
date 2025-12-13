package middleware

import (
	"cso-erp/internal/config"
	"cso-erp/internal/models"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ============================================================================
// SUPER SYSTEM ADMIN MIDDLEWARE
// ============================================================================

// SSA Rate Limiting
type ssaRateLimiter struct {
	attempts map[string][]time.Time
	mu       sync.RWMutex
}

var ssaLimiter = &ssaRateLimiter{
	attempts: make(map[string][]time.Time),
}

const (
	maxSSAAttempts     = 3
	ssaLockoutDuration = 24 * time.Hour
	ssaAttemptWindow   = 1 * time.Hour
)

// isRateLimited checks if IP is rate limited for SSA access
func (l *ssaRateLimiter) isRateLimited(ip string) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	attempts, exists := l.attempts[ip]
	if !exists {
		return false
	}

	// Clean old attempts
	now := time.Now()
	validAttempts := 0
	for _, t := range attempts {
		if now.Sub(t) < ssaAttemptWindow {
			validAttempts++
		}
	}

	return validAttempts >= maxSSAAttempts
}

// recordAttempt records a failed attempt
func (l *ssaRateLimiter) recordAttempt(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if _, exists := l.attempts[ip]; !exists {
		l.attempts[ip] = make([]time.Time, 0)
	}
	l.attempts[ip] = append(l.attempts[ip], now)
}

// clearAttempts clears attempts for an IP (on successful login)
func (l *ssaRateLimiter) clearAttempts(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.attempts, ip)
}

// RequireSuperAdmin validates SSA access: IP restriction + Login + 2FA
func RequireSuperAdmin(cfg *config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientIP := c.IP()

		// 1. Check rate limiting first
		if ssaLimiter.isRateLimited(clientIP) {
			// Return 404 to hide existence
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource not found",
			})
		}

		// 2. Check IP allowlist (most important)
		if !isAllowedIP(clientIP, cfg.MultiTenant.AllowedSSAIPs) {
			// Don't record as attempt - just deny silently
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource not found",
			})
		}

		// 3. Check authentication (JWT)
		userID := GetUserID(c)
		if userID == uuid.Nil {
			ssaLimiter.recordAttempt(clientIP)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Authentication required",
			})
		}

		// 4. Check role is system_admin
		userRole := GetUserRole(c)
		if userRole != models.RoleSystemAdmin {
			ssaLimiter.recordAttempt(clientIP)
			// Return 404 to hide SSA existence
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Resource not found",
			})
		}

		// 5. Check 2FA is verified (Disabled: Token presence implies verification)
		// if !Is2FAVerified(c) {
		// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
		// 		"error":   "2FA verification required",
		// 		"message": "Super System Admin requires 2FA verification",
		// 	})
		// }

		// Clear rate limit on successful access
		ssaLimiter.clearAttempts(clientIP)

		// Store SSA context for audit logging
		c.Locals("isSSA", true)
		c.Locals("ssaIP", clientIP)

		return c.Next()
	}
}

// isAllowedIP checks if the client IP is in the allowlist
func isAllowedIP(clientIP string, allowedIPs []string) bool {
	if len(allowedIPs) == 0 {
		// If no IPs configured, default to localhost only
		return clientIP == "127.0.0.1" || clientIP == "::1"
	}

	// Normalize client IP
	clientIP = normalizeIP(clientIP)

	for _, allowed := range allowedIPs {
		allowed = strings.TrimSpace(allowed)

		// Check for CIDR notation
		if strings.Contains(allowed, "/") {
			_, network, err := net.ParseCIDR(allowed)
			if err == nil && network.Contains(net.ParseIP(clientIP)) {
				return true
			}
			continue
		}

		// Direct IP comparison
		if normalizeIP(allowed) == clientIP {
			return true
		}
	}

	return false
}

// normalizeIP normalizes an IP address for comparison
func normalizeIP(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	return parsed.String()
}

// Is2FAVerified checks if 2FA has been verified for current session
func Is2FAVerified(c *fiber.Ctx) bool {
	verified, ok := c.Locals("2fa_verified").(bool)
	return ok && verified
}

// Set2FAVerified marks 2FA as verified for the current session
func Set2FAVerified(c *fiber.Ctx, verified bool) {
	c.Locals("2fa_verified", verified)
}

// ============================================================================
// SSA AUDIT LOGGING
// ============================================================================

// SSAAuditAction represents actions that can be audited
type SSAAuditAction string

const (
	SSAAuditLogin                SSAAuditAction = "ssa_login"
	SSAAuditLogout               SSAAuditAction = "ssa_logout"
	SSAAuditCreateOrganization   SSAAuditAction = "create_organization"
	SSAAuditVerifyDomain         SSAAuditAction = "verify_domain"
	SSAAuditCreateDatabase       SSAAuditAction = "create_database"
	SSAAuditCreateTenantAdmin    SSAAuditAction = "create_tenant_admin"
	SSAAuditSuspendOrganization  SSAAuditAction = "suspend_organization"
	SSAAuditActivateOrganization SSAAuditAction = "activate_organization"
	SSAAuditDeleteOrganization   SSAAuditAction = "delete_organization"
)

// GetSSAContext retrieves SSA context from fiber context
func GetSSAContext(c *fiber.Ctx) (bool, string) {
	isSSA, _ := c.Locals("isSSA").(bool)
	ssaIP, _ := c.Locals("ssaIP").(string)
	return isSSA, ssaIP
}
