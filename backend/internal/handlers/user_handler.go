package handlers

import (
	"context"
	"cso-erp/internal/config"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type UserHandler struct {
	cfg          *config.Config
	emailService *services.EmailService // Master email service
}

func NewUserHandler(cfg *config.Config, emailService *services.EmailService) *UserHandler {
	return &UserHandler{
		cfg:          cfg,
		emailService: emailService,
	}
}

// getServices creates request-scoped services with tenant DB
func (h *UserHandler) getServices(c *fiber.Ctx) (*services.UserService, *services.EmailService, error) {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil, nil, fiber.NewError(fiber.StatusInternalServerError, "Database connection not found")
	}

	userService := services.NewUserService(db, h.cfg)

	// Setup SMTP Service for this tenant to get correct EmailService
	smtpService := services.NewSMTPSettingsService(db, h.cfg)
	smtpSettings, err := smtpService.GetDefaultSMTPSettings(c.Context())

	var tenantEmailService *services.EmailService
	if err == nil && smtpSettings != nil {
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

	finalEmailService := services.GetHierarchicalEmailService(tenantEmailService, h.emailService)

	return userService, finalEmailService, nil
}

func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	userService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	users, err := userService.ListUsers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	return c.JSON(users)
}

func (h *UserHandler) GetUser(c *fiber.Ctx) error {
	userService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	user, err := userService.GetUserByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(user)
}

func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	userService, emailService, err := h.getServices(c)
	if err != nil {
		return err
	}

	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username, email, and password are required",
		})
	}

	currentUserID := middleware.GetUserID(c)

	user, err := userService.CreateUser(c.Context(), &req, currentUserID)
	if err != nil {
		log.Printf("[CreateUser] Error creating user '%s': %v", req.Username, err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Send welcome email (non-blocking, in a goroutine)
	// Important: We pass the resolved emailService which might be tenant specific
	go func(es *services.EmailService, u *models.User) {
		fullName := u.FullName
		if u.FullNameAr != nil && *u.FullNameAr != "" {
			fullName = *u.FullNameAr
		}
		err := es.SendWelcomeEmail(context.Background(), u.Email, fullName, u.Username, string(u.Role))
		if err != nil {
			log.Printf("Failed to send welcome email to %s: %v", u.Email, err)
		} else {
			log.Printf("Welcome email sent to %s", u.Email)
		}
	}(emailService, user)

	return c.Status(fiber.StatusCreated).JSON(user)
}

func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	userService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req models.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := userService.UpdateUser(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(user)
}

func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	userService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	if err := userService.DeleteUser(c.Context(), id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "User deleted successfully",
	})
}

func (h *UserHandler) UpdateUserRole(c *fiber.Ctx) error {
	userService, _, err := h.getServices(c)
	if err != nil {
		return err
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		Role models.UserRole `json:"role"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	user, err := userService.UpdateUserRole(c.Context(), id, req.Role)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(user)
}
