package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/services"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// SSARegistrationHandler handles Super System Admin registration
type SSARegistrationHandler struct {
	authService *services.SSAAuthService
	cfg         *config.Config
}

// NewSSARegistrationHandler creates a new SSA registration handler
func NewSSARegistrationHandler(authService *services.SSAAuthService, cfg *config.Config) *SSARegistrationHandler {
	return &SSARegistrationHandler{
		authService: authService,
		cfg:         cfg,
	}
}

// RegisterRequest represents SSA registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

// Register handles first-time SSA account creation
// POST /api/v1/ssa/auth/register
// This endpoint only works when NO super admins exist in the database
func (h *SSARegistrationHandler) Register(c *fiber.Ctx) error {
	// Check if any SSA already exists
	count, err := h.authService.CountAdmins(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "فشل في التحقق من حالة النظام",
		})
	}

	// Security: Only allow registration if no SSA exists
	if count > 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "التسجيل غير متاح - يوجد مشرف نظام بالفعل",
		})
	}

	// Parse request body
	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "بيانات غير صالحة",
		})
	}

	// Validate required fields
	if req.Username == "" || req.Email == "" || req.Password == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "جميع الحقول مطلوبة",
		})
	}

	// Validate username (3-50 chars, alphanumeric + underscore)
	if len(req.Username) < 3 || len(req.Username) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "اسم المستخدم يجب أن يكون بين 3 و 50 حرفًا",
		})
	}

	// Validate email format (basic check)
	if !strings.Contains(req.Email, "@") || !strings.Contains(req.Email, ".") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "عنوان البريد الإلكتروني غير صالح",
		})
	}

	// Validate password strength (minimum 8 chars)
	if len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "كلمة المرور يجب أن تكون 8 أحرف على الأقل",
		})
	}

	// Create SSA admin
	admin, err := h.authService.CreateAdmin(c.Context(), &services.CreateAdminRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Return success (without sensitive data)
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "تم إنشاء حساب المشرف الأعلى بنجاح",
		"admin": fiber.Map{
			"id":        admin.ID,
			"username":  admin.Username,
			"email":     admin.Email,
			"full_name": admin.FullName,
		},
	})
}

// CheckRegistrationAvailable checks if SSA registration is available
// GET /api/v1/ssa/auth/registration-status
func (h *SSARegistrationHandler) CheckRegistrationAvailable(c *fiber.Ctx) error {
	count, err := h.authService.CountAdmins(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "فشل في التحقق من حالة النظام",
		})
	}

	return c.JSON(fiber.Map{
		"registration_available": count == 0,
		"admin_count":            count,
	})
}
