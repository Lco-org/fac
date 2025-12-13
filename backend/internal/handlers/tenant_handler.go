package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TenantHandler handles multi-tenant operations for Super System Admin
type TenantHandler struct {
	tenantService *services.TenantService
	emailService  *services.EmailService
	cfg           *config.Config
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *services.TenantService, emailService *services.EmailService, cfg *config.Config) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		emailService:  emailService,
		cfg:           cfg,
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

	// Send verification email
	if h.emailService != nil {
		err = h.emailService.SendVerificationCode(req.ContactEmail, org.Name, code)
		if err != nil {
			// Log error but don't fail - code was generated
			// log.Printf("Failed to send verification email: %v", err)
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

	// Send verification email
	if h.emailService != nil {
		org, _ := h.tenantService.GetOrganizationByID(c.Context(), orgID)
		orgName := "Organization"
		if org != nil {
			orgName = org.Name
		}
		_ = h.emailService.SendVerificationCode(req.Email, orgName, code)
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
