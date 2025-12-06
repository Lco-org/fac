package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/security"
	"cso-erp/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *services.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *services.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username, email, and password are required",
		})
	}

	nilUUID := uuid.Nil

	user, err := h.authService.CreateUser(
		c.Context(),
		req.Username,
		req.Email,
		req.Password,
		req.FullName,
		req.Role,
		nilUUID,
	)
	if err != nil {
		switch err {
		case services.ErrEmailDomainNotAllowed:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Only @lco.org.ly email addresses are allowed",
			})
		case services.ErrWeakPassword:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Password must be at least 8 characters",
			})
		default:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Registration failed",
			})
		}
	}

	username := user.Username
	h.logAuditAction(c, security.ActionUserCreate, "user", &user.ID, &username, security.AuditDetails(map[string]interface{}{
		"email": user.Email,
		"role":  user.Role,
	}))

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    user,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	result, err := h.authService.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		h.logAuditAction(c, security.ActionLoginFailed, "auth", nil, &req.Username, security.AuditDetails(map[string]interface{}{
			"reason": "Invalid credentials",
		}))

		if err == services.ErrInvalidCredentials || err == services.ErrUserNotActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid credentials",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Authentication failed",
		})
	}

	if result.Requires2FA {
		return c.JSON(fiber.Map{
			"requires_2fa":    true,
			"pending_user_id": result.PendingUserID.String(),
			"message":         "تم إرسال رمز التحقق إلى بريدك الإلكتروني",
		})
	}

	return h.completeLogin(c, result.User)
}

func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"user_id"`
		Code   string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.UserID == "" || req.Code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "User ID and verification code are required",
		})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user, err := h.authService.Verify2FA(c.Context(), userID, req.Code)
	if err != nil {
		switch err {
		case security.ErrOTPExpired:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "رمز التحقق منتهي الصلاحية، يرجى طلب رمز جديد",
			})
		case security.ErrOTPInvalid:
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "رمز التحقق غير صحيح",
			})
		case security.ErrTooManyAttempts:
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "محاولات كثيرة جداً، يرجى طلب رمز جديد",
			})
		case security.ErrOTPNotFound:
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "لم يتم العثور على رمز تحقق، يرجى تسجيل الدخول مجدداً",
			})
		default:
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "فشل التحقق",
			})
		}
	}

	return h.completeLogin(c, user)
}

func (h *AuthHandler) ResendOTP(c *fiber.Ctx) error {
	var req struct {
		UserID string `json:"user_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	err = h.authService.ResendOTP(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send verification code",
		})
	}

	return c.JSON(fiber.Map{
		"message": "تم إرسال رمز تحقق جديد إلى بريدك الإلكتروني",
	})
}

func (h *AuthHandler) completeLogin(c *fiber.Ctx, user *models.User) error {
	token, err := middleware.GenerateToken(user, &h.cfg.JWT)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Session creation failed",
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(h.cfg.JWT.Expiry),
		HTTPOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: "Strict",
		Path:     "/",
	})

	username := user.Username
	h.logAuditAction(c, security.ActionLogin, "auth", &user.ID, &username, security.AuditDetails(map[string]interface{}{
		"role": user.Role,
	}))

	return c.JSON(models.LoginResponse{
		User:  *user,
		Token: token,
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	if userID != uuid.Nil {
		h.logAuditAction(c, security.ActionLogout, "auth", &userID, &username, "{}")
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: "Strict",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)

	user, err := h.authService.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func (h *AuthHandler) RevokeAllSessions(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	err := h.authService.RevokeAllSessions(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke sessions",
		})
	}

	h.logAuditAction(c, security.ActionRevokeAllSessions, "auth", &userID, &username, security.AuditDetails(map[string]interface{}{
		"action": "All sessions revoked",
	}))

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: "Strict",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "All sessions have been revoked. Please login again.",
	})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	var req struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Old and new passwords are required",
		})
	}

	err := h.authService.ChangePassword(c.Context(), userID, req.OldPassword, req.NewPassword)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Current password is incorrect",
			})
		}
		if err == services.ErrWeakPassword {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "New password must be at least 8 characters",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Password change failed",
		})
	}

	h.logAuditAction(c, security.ActionPasswordChange, "auth", &userID, &username, "{}")

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: "Strict",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Password changed successfully. Please login again.",
	})
}

func (h *AuthHandler) Toggle2FA(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	var req struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	err := h.authService.Toggle2FA(c.Context(), userID, req.Enabled)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update 2FA settings",
		})
	}

	h.logAuditAction(c, security.ActionDataModified, "user", &userID, &username, security.AuditDetails(map[string]interface{}{
		"two_factor_enabled": req.Enabled,
	}))

	status := "disabled"
	if req.Enabled {
		status = "enabled"
	}

	return c.JSON(fiber.Map{
		"message":            "Two-factor authentication " + status,
		"two_factor_enabled": req.Enabled,
	})
}

func (h *AuthHandler) logAuditAction(c *fiber.Ctx, action security.AuditAction, resourceType string, resourceID *uuid.UUID, username *string, details string) {
	entry := &security.AuditLog{
		Action:       action,
		UserID:       nil,
		Username:     username,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    middleware.GetClientIP(c),
		UserAgent:    middleware.GetUserAgent(c),
	}

	if userID := middleware.GetUserID(c); userID != uuid.Nil {
		entry.UserID = &userID
	}

	go func() {
		_ = h.authService.GetAuditService().LogAction(c.Context(), entry)
	}()
}
