package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// EmailTemplateHandler handles email template API endpoints
type EmailTemplateHandler struct {
	cfg *config.Config
}

// NewEmailTemplateHandler creates a new email template handler
func NewEmailTemplateHandler(cfg *config.Config) *EmailTemplateHandler {
	return &EmailTemplateHandler{
		cfg: cfg,
	}
}

// getServices creates request-scoped services with tenant DB
func (h *EmailTemplateHandler) getServices(c *fiber.Ctx) (*services.EmailTemplateService, *services.SMTPSettingsService, error) {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "Database connection not found")
	}

	templateService := services.NewEmailTemplateService(db)
	smtpService := services.NewSMTPSettingsService(db, h.cfg)

	return templateService, smtpService, nil
}

// ============================================================================
// EMAIL TEMPLATE ENDPOINTS
// ============================================================================

// ListTemplates returns all email templates
// GET /api/v1/organization/email-templates
func (h *EmailTemplateHandler) ListTemplates(c *fiber.Ctx) error {
	templateService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	templates, err := templateService.ListTemplates(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"templates": templates,
		"count":     len(templates),
	})
}

// GetTemplate returns a single template by ID
// GET /api/v1/organization/email-templates/:id
func (h *EmailTemplateHandler) GetTemplate(c *fiber.Ctx) error {
	templateService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid template ID",
		})
	}

	template, err := templateService.GetTemplateByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if template == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Template not found",
		})
	}

	return c.JSON(template)
}

// UpdateTemplate updates an email template
// PUT /api/v1/organization/email-templates/:id
func (h *EmailTemplateHandler) UpdateTemplate(c *fiber.Ctx) error {
	templateService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid template ID",
		})
	}

	var req models.UpdateEmailTemplateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	template, err := templateService.UpdateTemplate(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Template updated successfully",
		"template": template,
	})
}

// PreviewTemplate previews a template with variables
// POST /api/v1/organization/email-templates/preview
func (h *EmailTemplateHandler) PreviewTemplate(c *fiber.Ctx) error {
	templateService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	var req models.TemplatePreviewRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get language preference from header or query
	useArabic := c.Query("lang") == "ar" || c.Get("Accept-Language") == "ar"

	preview, err := templateService.RenderTemplate(c.Context(), req.TemplateKey, req.Variables, useArabic)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(preview)
}

// ResetTemplate resets a template to default
// POST /api/v1/organization/email-templates/:key/reset
func (h *EmailTemplateHandler) ResetTemplate(c *fiber.Ctx) error {
	templateService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	templateKey := c.Params("key")
	if templateKey == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Template key is required",
		})
	}

	err = templateService.ResetTemplate(c.Context(), templateKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Template reset to default",
	})
}

// ============================================================================
// SMTP SETTINGS ENDPOINTS
// ============================================================================

// ListSMTPSettings returns all SMTP configurations
// GET /api/v1/organization/smtp-settings
func (h *EmailTemplateHandler) ListSMTPSettings(c *fiber.Ctx) error {
	_, smtpService, err := h.getServices(c)
	if err != nil {
		return err
	}

	settings, err := smtpService.ListSMTPSettings(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"smtp_settings": settings,
		"count":         len(settings),
	})
}

// GetSMTPSettings returns SMTP settings by ID
// GET /api/v1/organization/smtp-settings/:id
func (h *EmailTemplateHandler) GetSMTPSettings(c *fiber.Ctx) error {
	_, smtpService, err := h.getServices(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SMTP settings ID",
		})
	}

	settings, err := smtpService.GetSMTPSettingsByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if settings == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "SMTP settings not found",
		})
	}

	return c.JSON(settings)
}

// CreateSMTPSettings creates new SMTP configuration
// POST /api/v1/organization/smtp-settings
func (h *EmailTemplateHandler) CreateSMTPSettings(c *fiber.Ctx) error {
	_, smtpService, err := h.getServices(c)
	if err != nil {
		return err
	}

	var req models.CreateSMTPSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := middleware.GetUserID(c)

	settings, err := smtpService.CreateSMTPSettings(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":       "SMTP settings created successfully",
		"smtp_settings": settings,
	})
}

// UpdateSMTPSettings updates SMTP configuration
// PUT /api/v1/organization/smtp-settings/:id
func (h *EmailTemplateHandler) UpdateSMTPSettings(c *fiber.Ctx) error {
	_, smtpService, err := h.getServices(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SMTP settings ID",
		})
	}

	var req models.UpdateSMTPSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	settings, err := smtpService.UpdateSMTPSettings(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":       "SMTP settings updated successfully",
		"smtp_settings": settings,
	})
}

// DeleteSMTPSettings soft-deletes SMTP settings
// DELETE /api/v1/organization/smtp-settings/:id
func (h *EmailTemplateHandler) DeleteSMTPSettings(c *fiber.Ctx) error {
	_, smtpService, err := h.getServices(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SMTP settings ID",
		})
	}

	err = smtpService.DeleteSMTPSettings(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "SMTP settings deleted successfully",
	})
}

// TestSMTPSettings tests SMTP connection by sending a test email
// POST /api/v1/organization/smtp-settings/:id/test
func (h *EmailTemplateHandler) TestSMTPSettings(c *fiber.Ctx) error {
	_, smtpService, err := h.getServices(c)
	if err != nil {
		return err
	}

	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SMTP settings ID",
		})
	}

	var req models.TestSMTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get SMTP settings
	settings, err := smtpService.GetSMTPSettingsByID(c.Context(), id)
	if err != nil || settings == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "SMTP settings not found",
		})
	}

	// Get decrypted password
	password, err := smtpService.GetDecryptedPassword(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve SMTP password",
		})
	}

	// Configure temporary email service
	fromName := ""
	if settings.FromName != nil {
		fromName = *settings.FromName
	}

	tenantConfig := &services.TenantEmailConfig{
		SMTPHost:    settings.SMTPHost,
		SMTPPort:    settings.SMTPPort,
		Username:    settings.Username,
		Password:    password,
		FromAddress: settings.FromAddress,
		FromName:    fromName,
		UseTLS:      settings.UseTLS,
	}

	tempEmailService := services.NewEmailServiceWithTenantConfig(tenantConfig)
	if tempEmailService == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid SMTP configuration",
		})
	}

	// Send test email
	subject := "Test Email - CSO ERP SMTP Settings"
	body := fmt.Sprintf("<h1>SMTP Test</h1><p>This is a test email to verify SMTP settings for %s.</p>", settings.SMTPHost)

	err = tempEmailService.SendEmail(c.Context(), req.ToEmail, subject, body, true)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to send test email: " + err.Error(),
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Test email sent successfully",
		"smtp_host":    settings.SMTPHost,
		"smtp_port":    settings.SMTPPort,
		"to_email":     req.ToEmail,
		"test_success": true,
	})
}
