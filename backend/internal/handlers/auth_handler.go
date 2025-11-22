package handlers

import (
	"cso-erp/internal/config"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
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

// Register handles user registration
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate input
	if req.Username == "" || req.Password == "" || req.Email == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username, email, and password are required",
		})
	}

	// For registration, we don't have a createdBy user, so use nil UUID
	nilUUID := uuid.Nil

	// Create user
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"user":    user,
	})
}

// Login handles user authentication
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	// Authenticate user
	user, err := h.authService.Login(c.Context(), req.Username, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials || err == services.ErrUserNotActive {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal server error",
		})
	}

	// Generate JWT token
	token, err := middleware.GenerateToken(user, &h.cfg.JWT)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// Set HttpOnly cookie (XSS protection)
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(h.cfg.JWT.Expiry),
		HTTPOnly: true,
		Secure:   h.cfg.App.Env == "production",
		SameSite: "Lax",
	})

	return c.JSON(models.LoginResponse{
		User:  *user,
		Token: token,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Clear cookie
	c.Cookie(&fiber.Cookie{
		Name:     "auth_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HTTPOnly: true,
	})

	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

// GetCurrentUser returns current authenticated user
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
