package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/services"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// SSAAuthHandler handles Super System Admin authentication
type SSAAuthHandler struct {
	authService *services.SSAAuthService
	cfg         *config.Config
}

// NewSSAAuthHandler creates a new SSA auth handler
func NewSSAAuthHandler(authService *services.SSAAuthService, cfg *config.Config) *SSAAuthHandler {
	return &SSAAuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// getClientIP extracts client IP from request
func (h *SSAAuthHandler) getClientIP(c *fiber.Ctx) string {
	// Check X-Forwarded-For first (for proxies/load balancers)
	xff := c.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}
	// Check X-Real-IP
	xrip := c.Get("X-Real-IP")
	if xrip != "" {
		return xrip
	}
	// Fall back to direct IP
	return c.IP()
}

// isIPAllowed checks if client IP is in allowed list
func (h *SSAAuthHandler) isIPAllowed(clientIP string) bool {
	allowedIPs := h.cfg.MultiTenant.AllowedSSAIPs
	if len(allowedIPs) == 0 {
		return true // No restriction if not configured
	}

	for _, ip := range allowedIPs {
		if ip == clientIP {
			return true
		}
	}
	return false
}

// Login handles SSA login with username and password
// POST /api/v1/ssa/auth/login
func (h *SSAAuthHandler) Login(c *fiber.Ctx) error {
	// Check IP restriction
	clientIP := h.getClientIP(c)

	// Log for debugging
	log.Printf("SSA Login attempt from IP: %s", clientIP)

	if !h.isIPAllowed(clientIP) {
		log.Printf("SSA Login: IP not allowed: %s", clientIP)
		// Return 404 to hide SSA existence
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	}

	var req services.SSALoginRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("SSA Login: Bad request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "بيانات غير صالحة",
		})
	}

	log.Printf("SSA Login: Attempting login for user: %s", req.Username)

	// Validate
	if req.Username == "" || req.Password == "" {
		log.Printf("SSA Login: Missing username or password")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "اسم المستخدم وكلمة المرور مطلوبان",
		})
	}

	// Attempt login
	result, err := h.authService.Login(c.Context(), &req, clientIP)
	if err != nil {
		log.Printf("SSA Login failed for %s: %v", req.Username, err)
		// Log for debugging
		c.Response().Header.Set("X-SSA-Login-Error", err.Error())
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Handle No-2FA login
	if !result.Requires2FA && result.Token != "" {
		// Set auth cookie
		c.Cookie(&fiber.Cookie{
			Name:     "auth_token",
			Value:    result.Token,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour), // 24 hours
			HTTPOnly: true,
			Secure:   h.cfg.Security.CookieSecure,
			SameSite: "Lax",
		})
	}

	return c.JSON(fiber.Map{
		"requires_2fa":       result.Requires2FA,
		"pending_session_id": result.PendingSessionID,
		"message":            result.Message,
		"token":              result.Token,
		"admin":              result.Admin,
	})
}

// Verify2FA handles 2FA code verification
// POST /api/v1/ssa/auth/verify-2fa
func (h *SSAAuthHandler) Verify2FA(c *fiber.Ctx) error {
	// Check IP restriction
	clientIP := h.getClientIP(c)
	if !h.isIPAllowed(clientIP) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Not Found",
		})
	}

	var req struct {
		SessionID string `json:"session_id"`
		Code      string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "بيانات غير صالحة",
		})
	}

	// Validate
	if req.SessionID == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "رمز التحقق مطلوب",
		})
	}

	// Verify 2FA
	result, err := h.authService.Verify2FA(c.Context(), req.SessionID, req.Code, clientIP)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Set auth cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    result.Token,
		Path:     "/",
		MaxAge:   14400, // 4 hours
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{
		"success": true,
		"admin":   result.Admin,
		"token":   result.Token,
		"message": result.Message,
	})
}

// GetMe returns current SSA user
func (h *SSAAuthHandler) GetMe(c *fiber.Ctx) error {
	// 1. Get UserID from Locals (set by AuthMiddleware)
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Unauthorized"})
	}

	// 2. Call Service
	admin, err := h.authService.GetAdminByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}

	// Return map with explicit role for frontend
	return c.JSON(fiber.Map{
		"id":                 admin.ID,
		"username":           admin.Username,
		"email":              admin.Email,
		"full_name":          admin.FullName,
		"role":               "system_admin", // Required for frontend router guards
		"two_factor_enabled": admin.TwoFactorEnabled,
		"is_active":          admin.IsActive,
		"created_at":         admin.CreatedAt,
	})
}

// Logout handles SSA logout
// POST /api/v1/ssa/auth/logout
func (h *SSAAuthHandler) Logout(c *fiber.Ctx) error {
	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Secure:   true,
		HTTPOnly: true,
		SameSite: "Strict",
	})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "تم تسجيل الخروج بنجاح",
	})
}

// GetCurrentAdmin returns current SSA admin info
// GET /api/v1/ssa/auth/me
func (h *SSAAuthHandler) GetCurrentAdmin(c *fiber.Ctx) error {
	// This requires SSA auth middleware to be in place
	// For now, return from context (set by middleware)
	adminID := c.Locals("adminID")
	if adminID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "غير مصرح",
		})
	}

	return c.JSON(fiber.Map{
		"admin_id": adminID,
	})
}
