package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/security"
	"cso-erp/internal/services"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type AuthHandler struct {
	masterDB     *database.DB
	cfg          *config.Config
	emailService *services.EmailService // Master email service
}

func NewAuthHandler(masterDB *database.DB, cfg *config.Config, emailService *services.EmailService) *AuthHandler {
	return &AuthHandler{
		masterDB:     masterDB,
		cfg:          cfg,
		emailService: emailService,
	}
}

// getTenantService creates a request-scoped AuthService with tenant DB and SMTP
func (h *AuthHandler) getTenantService(c *fiber.Ctx) (*services.AuthService, error) {
	var db *database.DB

	// Priority 1: Check for tenant_db from JWT claims (set by AuthMiddleware)
	// This ensures authenticated users always use their correct tenant database
	if tenantDBName := middleware.GetTenantDBName(c); tenantDBName != "" {
		tenantConfig := &config.DatabaseConfig{
			Host:     h.cfg.Database.Host,
			Port:     h.cfg.Database.Port,
			User:     h.cfg.Database.User,
			Password: h.cfg.Database.Password,
			DBName:   tenantDBName,
			SSLMode:  h.cfg.Database.SSLMode,
		}

		var err error
		db, err = database.Connect(tenantConfig)
		if err != nil {
			log.Printf("[getTenantService] Failed to connect to JWT tenant DB '%s': %v", tenantDBName, err)
		} else {
			log.Printf("[getTenantService] Using tenant DB from JWT: %s", tenantDBName)
		}
	}

	// Priority 2: Fall back to TenantResolver's context (for unauthenticated routes like login)
	if db == nil {
		db = middleware.GetTenantDB(c)
	}

	if db == nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "Database connection not found")
	}

	// Setup SMTP Service for this tenant
	smtpService := services.NewSMTPSettingsService(db, h.cfg)

	// Get default SMTP settings
	// We use background context as we don't want to cancel if request cancels (maybe?)
	// Actually, using request context is fine.
	smtpSettings, err := smtpService.GetDefaultSMTPSettings(c.Context())

	var tenantEmailService *services.EmailService
	if err == nil && smtpSettings != nil {
		// Decrypt password
		password, _ := smtpService.GetDecryptedPassword(c.Context(), smtpSettings.ID)

		fromName := ""
		if smtpSettings.FromName != nil {
			fromName = *smtpSettings.FromName
		}

		tenantConfig := &services.TenantEmailConfig{
			SMTPHost:    smtpSettings.SMTPHost,
			SMTPPort:    smtpSettings.SMTPPort,
			Username:    smtpSettings.Username,
			Password:    password,
			FromAddress: smtpSettings.FromAddress,
			FromName:    fromName,
			UseTLS:      smtpSettings.UseTLS,
		}
		tenantEmailService = services.NewEmailServiceWithTenantConfig(tenantConfig)
	}

	// Use hierarchical email service (Tenant > Master)
	finalEmailService := services.GetHierarchicalEmailService(tenantEmailService, h.emailService)

	// Create AuthService with Tenant DB and correct Email Service
	return services.NewAuthService(db, h.cfg, finalEmailService), nil
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	svc, err := h.getTenantService(c)
	if err != nil {
		return err
	}

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

	user, err := svc.CreateUser(
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
	h.logAuditAction(c, svc.GetAuditService(), security.ActionUserCreate, "user", &user.ID, &username, security.AuditDetails(map[string]interface{}{
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
	// Parse body first to check email domain if needed
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

	// SMART LOGIN: Always try email domain lookup FIRST
	// This ensures users are directed to their correct tenant database regardless of URL
	var svc *services.AuthService

	email := req.Username
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		domain := parts[1]

		// Lookup tenant by domain in MasterDB
		var dbName string
		lookupErr := h.masterDB.QueryRowContext(c.Context(), `
			SELECT database_name FROM organizations 
			WHERE (primary_domain = $1 OR id IN (SELECT organization_id FROM allowed_domains WHERE domain = $1))
			AND status = 'active'
		`, domain).Scan(&dbName)

		if lookupErr == nil && dbName != "" {
			// Found tenant by email domain! Connect to it
			tenantConfig := &config.DatabaseConfig{
				Host:     h.cfg.Database.Host,
				Port:     h.cfg.Database.Port,
				User:     h.cfg.Database.User,
				Password: h.cfg.Database.Password,
				DBName:   dbName,
				SSLMode:  h.cfg.Database.SSLMode,
			}

			if tenantDB, connErr := database.Connect(tenantConfig); connErr == nil {
				log.Printf("[SmartLogin] Matched email domain '%s' to database '%s'", domain, dbName)
				svc = services.NewAuthService(tenantDB, h.cfg, h.emailService)
				// Store tenant DB name for token generation
				c.Locals("login_tenant_db", dbName)
			}
		}
	}

	// Fallback: If no tenant found by email domain, use TenantResolver's default
	if svc == nil {
		var err error
		svc, err = h.getTenantService(c)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "خطأ في الاتصال بقاعدة البيانات. يرجى التأكد من الرابط المستخدم.",
			})
		}
	}

	result, err := svc.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		h.logAuditAction(c, svc.GetAuditService(), security.ActionLoginFailed, "auth", nil, &req.Username, security.AuditDetails(map[string]interface{}{
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

	return h.completeLogin(c, svc, result.User)
}

func (h *AuthHandler) Verify2FA(c *fiber.Ctx) error {
	svc, err := h.getTenantService(c)
	if err != nil {
		return err
	}

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

	user, err := svc.Verify2FA(c.Context(), userID, req.Code)
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

	return h.completeLogin(c, svc, user)
}

func (h *AuthHandler) ResendOTP(c *fiber.Ctx) error {
	svc, err := h.getTenantService(c)
	if err != nil {
		return err
	}

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

	err = svc.ResendOTP(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to send verification code",
		})
	}

	return c.JSON(fiber.Map{
		"message": "تم إرسال رمز تحقق جديد إلى بريدك الإلكتروني",
	})
}

func (h *AuthHandler) completeLogin(c *fiber.Ctx, svc *services.AuthService, user *models.User) error {
	// Get tenant DB from Smart Login path (stored in c.Locals during login)
	tenantDB := ""
	if tdb, ok := c.Locals("login_tenant_db").(string); ok && tdb != "" {
		tenantDB = tdb
	} else if tdb := middleware.GetTenantDBName(c); tdb != "" {
		// Fallback to tenant from TenantResolver if no Smart Login DB
		tenantDB = tdb
	}

	token, err := middleware.GenerateTokenWithTenant(user, &h.cfg.JWT, tenantDB)
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
		SameSite: "Lax",
		Path:     "/",
	})

	username := user.Username
	h.logAuditAction(c, svc.GetAuditService(), security.ActionLogin, "auth", &user.ID, &username, security.AuditDetails(map[string]interface{}{
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
		// Try to log audit, but don't fail if we can't get service (e.g. DB issue)
		if svc, err := h.getTenantService(c); err == nil {
			h.logAuditAction(c, svc.GetAuditService(), security.ActionLogout, "auth", &userID, &username, "{}")
		}
	}

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func (h *AuthHandler) GetCurrentUser(c *fiber.Ctx) error {
	svc, err := h.getTenantService(c)
	if err != nil {
		return err
	}

	userID := middleware.GetUserID(c)

	user, err := svc.GetUserByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func (h *AuthHandler) RevokeAllSessions(c *fiber.Ctx) error {
	svc, err := h.getTenantService(c)
	if err != nil {
		return err
	}

	userID := middleware.GetUserID(c)
	username := middleware.GetUsername(c)

	err = svc.RevokeAllSessions(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to revoke sessions",
		})
	}

	h.logAuditAction(c, svc.GetAuditService(), security.ActionRevokeAllSessions, "auth", &userID, &username, security.AuditDetails(map[string]interface{}{
		"action": "All sessions revoked",
	}))

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "All sessions have been revoked. Please login again.",
	})
}

func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	svc, err := h.getTenantService(c)
	if err != nil {
		return err
	}

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

	err = svc.ChangePassword(c.Context(), userID, req.OldPassword, req.NewPassword)
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

	h.logAuditAction(c, svc.GetAuditService(), security.ActionPasswordChange, "auth", &userID, &username, "{}")

	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Security.CookieSecure,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.JSON(fiber.Map{
		"message": "Password changed successfully. Please login again.",
	})
}

func (h *AuthHandler) Toggle2FA(c *fiber.Ctx) error {
	svc, err := h.getTenantService(c)
	if err != nil {
		return err
	}

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

	err = svc.Toggle2FA(c.Context(), userID, req.Enabled)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update 2FA settings",
		})
	}

	h.logAuditAction(c, svc.GetAuditService(), security.ActionDataModified, "user", &userID, &username, security.AuditDetails(map[string]interface{}{
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

func (h *AuthHandler) logAuditAction(c *fiber.Ctx, auditService *security.AuditService, action security.AuditAction, resourceType string, resourceID *uuid.UUID, username *string, details string) {
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
		_ = auditService.LogAction(c.Context(), entry)
	}()
}
