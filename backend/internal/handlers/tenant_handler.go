package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/security"
	"cso-erp/internal/services"
	"cso-erp/internal/validation"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantHandler handles multi-tenant operations for Super System Admin
type TenantHandler struct {
	tenantService  *services.TenantService
	emailService   *services.EmailService
	ssaAuthService *services.SSAAuthService
	cfg            *config.Config
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *services.TenantService, emailService *services.EmailService, ssaAuthService *services.SSAAuthService, cfg *config.Config) *TenantHandler {
	return &TenantHandler{
		tenantService:  tenantService,
		emailService:   emailService,
		ssaAuthService: ssaAuthService,
		cfg:            cfg,
	}
}

// ============================================================================
// ORGANIZATION ENDPOINTS
// ============================================================================

// CreateOrganization creates a new organization (SSA only)
// POST /api/ssa/organizations
func (h *TenantHandler) CreateOrganization(c *fiber.Ctx) error {
	var req models.CreateOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get SSA user ID from context
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Unauthorized",
		})
	}

	// Create organization
	org, err := h.tenantService.CreateOrganization(c.Context(), &req, adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Generate verification code
	code, err := h.tenantService.GenerateVerificationCode(c.Context(), org.ID, req.ContactEmail, c.IP())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Organization created but failed to generate verification code",
			"details": err.Error(),
		})
	}

	// Send verification email using master SMTP (from .env)
	if h.emailService != nil {
		// Log SSA email attempt
		if cfg := h.emailService.GetConfig(); cfg != nil {
			log.Printf("[SSA EMAIL] Sending verification to %s via %s (from: %s)", req.ContactEmail, cfg.SMTPHost, cfg.FromAddress)
		}

		err = h.emailService.SendVerificationCode(req.ContactEmail, org.Name, code)
		if err != nil {
			// Log error but don't fail - code was generated
			log.Printf("[SSA EMAIL ERROR] Failed to send verification email: %v", err)
		} else {
			log.Printf("[SSA EMAIL SUCCESS] Verification code sent to %s", req.ContactEmail)
		}
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":      "Organization created. Verification code sent to " + req.ContactEmail,
		"organization": org,
		"next_step":    "verify_domain",
	})
}

// ListOrganizations lists all organizations (SSA only)
// GET /api/ssa/organizations
func (h *TenantHandler) ListOrganizations(c *fiber.Ctx) error {
	status := c.Query("status", "")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	orgs, total, err := h.tenantService.ListOrganizations(c.Context(), status, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"organizations": orgs,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

// UpdateOrganizationSubscription updates an organization's subscription
// PUT /api/ssa/organizations/:id/subscription
func (h *TenantHandler) UpdateOrganizationSubscription(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid organization ID"})
	}

	var req models.UpdateOrganizationSubscriptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	v := validation.New()
	v.UUIDRequired("subscription_plan_id", req.SubscriptionPlanID)

	if v.HasErrors() {
		return c.Status(fiber.StatusBadRequest).JSON(v.Errors().ToMap())
	}

	err = h.tenantService.UpdateOrganizationSubscription(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"message": "Subscription updated successfully",
	})
}

// GetOrganization gets a single organization
// GET /api/ssa/organizations/:id
func (h *TenantHandler) GetOrganization(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	org, err := h.tenantService.GetOrganizationByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(org)
}

// ListSubscriptionPlans returns available subscription plans
// GET /api/ssa/subscription-plans
func (h *TenantHandler) ListSubscriptionPlans(c *fiber.Ctx) error {
	plans, err := h.tenantService.ListSubscriptionPlans(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to list plans",
		})
	}
	return c.JSON(fiber.Map{
		"plans": plans,
	})
}

// GetSubscriptionPlan returns a single subscription plan
// GET /api/ssa/subscription-plans/:id
func (h *TenantHandler) GetSubscriptionPlan(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid plan ID"})
	}

	plan, err := h.tenantService.GetSubscriptionPlan(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"plan": plan})
}

// CreateSubscriptionPlan creates a new subscription plan
// POST /api/ssa/subscription-plans
func (h *TenantHandler) CreateSubscriptionPlan(c *fiber.Ctx) error {
	var req models.CreateSubscriptionPlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	v := validation.New()
	v.Required("name", req.Name)
	v.Required("code", req.Code)
	v.AlphaNumeric("code", req.Code)
	v.Required("currency", req.Currency)
	v.LengthBetween("currency", req.Currency, 3, 3)
	v.NonNegativeFloat("price_monthly", req.PriceMonthly)
	v.NonNegativeFloat("price_yearly", req.PriceYearly)
	v.PositiveInt("max_users", req.MaxUsers)
	v.PositiveInt("max_storage_gb", req.MaxStorageGB)

	if v.HasErrors() {
		return c.Status(fiber.StatusBadRequest).JSON(v.Errors().ToMap())
	}

	plan, err := h.tenantService.CreateSubscriptionPlan(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"plan": plan})
}

// UpdateSubscriptionPlan updates a subscription plan
// PUT /api/ssa/subscription-plans/:id
func (h *TenantHandler) UpdateSubscriptionPlan(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid plan ID"})
	}

	var req models.UpdateSubscriptionPlanRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request body"})
	}

	plan, err := h.tenantService.UpdateSubscriptionPlan(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"plan": plan})
}

// DeleteSubscriptionPlan deletes a subscription plan
// DELETE /api/ssa/subscription-plans/:id
func (h *TenantHandler) DeleteSubscriptionPlan(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid plan ID"})
	}

	if err := h.tenantService.DeleteSubscriptionPlan(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "subscription plan deleted"})
}

// GetOrganizationStats gets dashboard statistics (SSA only)
// GET /api/ssa/stats
func (h *TenantHandler) GetOrganizationStats(c *fiber.Ctx) error {
	stats, err := h.tenantService.GetOrganizationStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// ============================================================================
// DOMAIN VERIFICATION ENDPOINTS
// ============================================================================

// VerifyDomain verifies domain ownership with code
// POST /api/ssa/organizations/:id/verify
func (h *TenantHandler) VerifyDomain(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.Code) != 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Verification code must be 6 characters",
		})
	}

	err = h.tenantService.VerifyCode(c.Context(), orgID, req.Code, c.IP())
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":   "Domain verified successfully",
		"next_step": "create_database",
	})
}

// ResendVerification resends verification code
// POST /api/ssa/organizations/:id/resend-verification
func (h *TenantHandler) ResendVerification(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Generate new code
	code, err := h.tenantService.GenerateVerificationCode(c.Context(), orgID, req.Email, c.IP())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Send verification email using master SMTP (from .env)
	if h.emailService != nil {
		org, _ := h.tenantService.GetOrganizationByID(c.Context(), orgID)
		orgName := "Organization"
		if org != nil {
			orgName = org.Name
		}

		// Log SSA email attempt
		if cfg := h.emailService.GetConfig(); cfg != nil {
			log.Printf("[SSA EMAIL] Sending verification to %s via %s (from: %s)", req.Email, cfg.SMTPHost, cfg.FromAddress)
		}

		if err := h.emailService.SendVerificationCode(req.Email, orgName, code); err != nil {
			log.Printf("[SSA EMAIL ERROR] Failed to send verification email: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to send email: " + err.Error(),
			})
		}
		log.Printf("[SSA EMAIL SUCCESS] Verification code sent to %s", req.Email)
	}

	return c.JSON(fiber.Map{
		"message": "Verification code resent to " + req.Email,
	})
}

// ============================================================================
// DATABASE CREATION ENDPOINTS
// ============================================================================

// CreateDatabase creates tenant database (SSA only)
// POST /api/ssa/organizations/:id/database
func (h *TenantHandler) CreateDatabase(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	// Get organization
	org, err := h.tenantService.GetOrganizationByID(c.Context(), orgID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Organization not found",
		})
	}

	// Check if verified
	if org.Status != models.OrgStatusActive {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Organization must be verified before creating database",
		})
	}

	// Create database
	err = h.tenantService.CreateTenantDatabase(c.Context(), org.DatabaseName)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":       "Database created successfully",
		"database_name": org.DatabaseName,
		"next_step":     "create_admin",
	})
}

// ============================================================================
// TENANT ADMIN ENDPOINTS
// ============================================================================

// CreateTenantAdmin creates the first admin for an organization
// POST /api/ssa/organizations/:id/admin
func (h *TenantHandler) CreateTenantAdmin(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	var req models.CreateTenantAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get organization
	org, err := h.tenantService.GetOrganizationByID(c.Context(), orgID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Organization not found",
		})
	}

	// Create tenant admin
	response, err := h.tenantService.CreateTenantAdmin(c.Context(), org, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Send welcome email with credentials
	if h.emailService != nil {
		_ = h.emailService.SendTenantWelcomeEmail(
			c.Context(),
			req.Email,
			req.FullName,
			org.Name,
			req.Username,
			response.TemporaryPassword,
			response.LoginURL,
		)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Tenant admin created successfully. Welcome email sent.",
		"admin":    response,
		"complete": true,
	})
}

// ============================================================================
// ORGANIZATION STATUS MANAGEMENT
// ============================================================================

// SuspendOrganization suspends an organization
// POST /api/ssa/organizations/:id/suspend
func (h *TenantHandler) SuspendOrganization(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	var req models.SuspendOrganizationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.tenantService.SuspendOrganization(c.Context(), orgID, req.Reason)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Organization suspended successfully",
	})
}

// ActivateOrganization activates a suspended organization
// POST /api/ssa/organizations/:id/activate
func (h *TenantHandler) ActivateOrganization(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid organization ID",
		})
	}

	err = h.tenantService.ActivateOrganization(c.Context(), orgID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Organization activated successfully",
	})
}

// ============================================================================
// ORGANIZATION DELETION (with OTP verification)
// ============================================================================

// RequestDeletion initiates organization deletion by sending OTP to admin's registered email
// Requires password verification before sending OTP
// POST /api/ssa/organizations/:id/request-deletion
func (h *TenantHandler) RequestDeletion(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "معرف المنظمة غير صالح",
		})
	}

	// Get admin ID from JWT context
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "غير مصرح",
		})
	}

	// Parse request - ONLY password required (email comes from database)
	var req struct {
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "بيانات غير صالحة",
		})
	}

	if req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "كلمة المرور مطلوبة للتحقق",
		})
	}

	// Get admin from database to verify password and get registered email
	admin, err := h.ssaAuthService.GetAdminByID(c.Context(), adminID)
	if err != nil || admin == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "لم يتم العثور على المدير",
		})
	}

	// Verify password using Argon2id
	validPassword, err := security.VerifyPassword(req.Password, admin.PasswordHash)
	if err != nil || !validPassword {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "كلمة المرور غير صحيحة",
		})
	}

	// Request deletion with OTP - using ONLY admin's registered email from DB
	result, err := h.tenantService.RequestOrganizationDeletion(
		c.Context(),
		orgID,
		adminID,
		admin.Email,    // ✅ Email from database, NOT user input
		admin.FullName, // ✅ Name from database, NOT user input
		c.IP(),
		h.emailService,
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Mask email for response
	maskedEmail := maskEmail(admin.Email)

	return c.JSON(fiber.Map{
		"success":      true,
		"message":      "تم إرسال رمز التحقق إلى بريدك الإلكتروني المسجل",
		"masked_email": maskedEmail,
		"request_id":   result.RequestID,
		"expires_at":   result.ExpiresAt,
	})
}

// maskEmail masks email for display (e.g., su***@lco.org.ly)
func maskEmail(email string) string {
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex < 2 {
		return "***" + email[atIndex:]
	}
	return email[:2] + "***" + email[atIndex:]
}

// ConfirmDeletion verifies OTP and permanently deletes organization
// POST /api/ssa/organizations/:id/confirm-deletion
func (h *TenantHandler) ConfirmDeletion(c *fiber.Ctx) error {
	idStr := c.Params("id")
	orgID, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "معرف المنظمة غير صالح",
		})
	}

	// Get admin info from context
	adminID := middleware.GetUserID(c)
	if adminID == uuid.Nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "غير مصرح",
		})
	}

	// Parse request body
	var req struct {
		Code string `json:"code"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "بيانات غير صالحة",
		})
	}

	if len(req.Code) != 6 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "رمز التحقق يجب أن يكون 6 أحرف",
		})
	}

	// Confirm deletion with OTP
	err = h.tenantService.ConfirmOrganizationDeletion(
		c.Context(),
		orgID,
		req.Code,
		adminID,
		c.IP(),
	)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "تم حذف المنظمة وجميع بياناتها نهائياً",
	})
}
