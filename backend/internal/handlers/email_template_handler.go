package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"fmt"
	"strings"

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
			"error":        "Invalid SMTP settings ID",
			"error_ar":     "معرف الإعدادات غير صالح",
			"error_type":   "INVALID_ID",
			"test_success": false,
		})
	}

	var req models.TestSMTPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        "Invalid request body",
			"error_ar":     "بيانات الطلب غير صالحة",
			"error_type":   "INVALID_REQUEST",
			"test_success": false,
		})
	}

	// Get SMTP settings
	settings, err := smtpService.GetSMTPSettingsByID(c.Context(), id)
	if err != nil || settings == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":        "SMTP settings not found",
			"error_ar":     "لم يتم العثور على إعدادات SMTP",
			"error_type":   "NOT_FOUND",
			"test_success": false,
		})
	}

	// Get decrypted password
	password, err := smtpService.GetDecryptedPassword(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":        "Failed to retrieve SMTP password",
			"error_ar":     "فشل في استرجاع كلمة مرور SMTP",
			"error_type":   "PASSWORD_ERROR",
			"test_success": false,
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
			"error":        "Invalid SMTP configuration",
			"error_ar":     "إعدادات SMTP غير صالحة - تحقق من عنوان البريد الإلكتروني",
			"error_type":   "INVALID_CONFIG",
			"test_success": false,
		})
	}

	// Send test email
	subject := "اختبار إعدادات البريد - CSO ERP"
	body := fmt.Sprintf(`<html dir="rtl"><body style="font-family: Arial; direction: rtl;">
		<h1 style="color: #22c55e;">✅ تم الاختبار بنجاح!</h1>
		<p>تم إرسال هذه الرسالة للتحقق من إعدادات SMTP.</p>
		<table style="border-collapse: collapse; margin-top: 20px;">
			<tr><td style="padding: 8px; border: 1px solid #ddd;"><strong>السيرفر:</strong></td><td style="padding: 8px; border: 1px solid #ddd;">%s</td></tr>
			<tr><td style="padding: 8px; border: 1px solid #ddd;"><strong>المنفذ:</strong></td><td style="padding: 8px; border: 1px solid #ddd;">%d</td></tr>
			<tr><td style="padding: 8px; border: 1px solid #ddd;"><strong>المرسل:</strong></td><td style="padding: 8px; border: 1px solid #ddd;">%s</td></tr>
		</table>
	</body></html>`, settings.SMTPHost, settings.SMTPPort, settings.FromAddress)

	err = tempEmailService.SendEmail(c.Context(), req.ToEmail, subject, body, true)
	if err != nil {
		errMsg := err.Error()
		errorType := "UNKNOWN"
		errorAr := "خطأ غير معروف"
		suggestion := ""

		// Categorize errors with Arabic explanations
		switch {
		case strings.Contains(errMsg, "no such host") || strings.Contains(errMsg, "lookup"):
			errorType = "DNS_ERROR"
			errorAr = "خطأ في اسم السيرفر"
			suggestion = "تحقق من صحة اسم السيرفر (SMTP Host)"
		case strings.Contains(errMsg, "connection refused"):
			errorType = "CONNECTION_REFUSED"
			errorAr = "تم رفض الاتصال"
			suggestion = "تحقق من رقم المنفذ - جرب 465 أو 587"
		case strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded"):
			errorType = "TIMEOUT"
			errorAr = "انتهت مهلة الاتصال"
			suggestion = "السيرفر لا يستجيب - تحقق من إعدادات الشبكة"
		case strings.Contains(errMsg, "535") || strings.Contains(errMsg, "Incorrect authentication"):
			errorType = "AUTH_ERROR"
			errorAr = "كلمة المرور خاطئة"
			suggestion = "تحقق من كلمة المرور - قد تحتاج لإنشاء كلمة مرور تطبيق"
		case strings.Contains(errMsg, "closed network") || strings.Contains(errMsg, "use of closed"):
			errorType = "CONNECTION_CLOSED"
			errorAr = "انقطع الاتصال بالسيرفر"
			suggestion = "السيرفر أغلق الاتصال - تحقق من بيانات الدخول أو جرب منفذ مختلف"
		case strings.Contains(errMsg, "authentication") || strings.Contains(errMsg, "auth"):
			errorType = "AUTH_ERROR"
			errorAr = "فشل تسجيل الدخول"
			suggestion = "تحقق من اسم المستخدم وكلمة المرور"
		case strings.Contains(errMsg, "TLS") || strings.Contains(errMsg, "certificate") || strings.Contains(errMsg, "STARTTLS"):
			errorType = "TLS_ERROR"
			errorAr = "خطأ في التشفير"
			suggestion = "جرب تغيير المنفذ: استخدم 465 مع SSL أو 587 مع TLS"
		case strings.Contains(errMsg, "relay") || strings.Contains(errMsg, "550") || strings.Contains(errMsg, "553"):
			errorType = "RELAY_ERROR"
			errorAr = "السيرفر لا يسمح بالإرسال"
			suggestion = "تحقق من صلاحيات حساب البريد"
		case strings.Contains(errMsg, "EOF") || strings.Contains(errMsg, "broken pipe"):
			errorType = "CONNECTION_LOST"
			errorAr = "فقد الاتصال بالسيرفر"
			suggestion = "حدثت مشكلة في الاتصال - أعد المحاولة"
		default:
			// For unknown errors, show the actual error message
			errorAr = "خطأ: " + errMsg
			suggestion = "راجع إعدادات SMTP أو تواصل مع مزود البريد"
		}

		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":        errMsg,
			"error_ar":     errorAr,
			"error_type":   errorType,
			"suggestion":   suggestion,
			"test_success": false,
			"smtp_host":    settings.SMTPHost,
			"smtp_port":    settings.SMTPPort,
		})
	}

	return c.JSON(fiber.Map{
		"message":      "Test email sent successfully",
		"message_ar":   "تم إرسال البريد التجريبي بنجاح ✅",
		"smtp_host":    settings.SMTPHost,
		"smtp_port":    settings.SMTPPort,
		"to_email":     req.ToEmail,
		"test_success": true,
	})
}
