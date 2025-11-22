package middleware

import (
	"cso-erp/internal/config"
	"cso-erp/internal/models"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JWTClaims struct {
	UserID   uuid.UUID       `json:"user_id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT token from HttpOnly cookie
func AuthMiddleware(cfg *config.JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from cookie (HttpOnly for XSS protection)
		tokenString := c.Cookies("auth_token")

		if tokenString == "" {
			// Fallback to Authorization header
			authHeader := c.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authentication token",
			})
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid token claims",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("user_role", claims.Role)

		return c.Next()
	}
}

// GenerateToken creates a new JWT token
func GenerateToken(user *models.User, cfg *config.JWTConfig) (string, error) {
	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) uuid.UUID {
	userID, _ := c.Locals("user_id").(uuid.UUID)
	return userID
}

// GetUserRole extracts user role from context
func GetUserRole(c *fiber.Ctx) models.UserRole {
	role, _ := c.Locals("user_role").(models.UserRole)
	return role
}
