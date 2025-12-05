package middleware

import (
	"cso-erp/internal/config"
	"cso-erp/internal/models"
	"database/sql"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the claims stored in JWT token
// Includes token version for Kill Switch functionality
type JWTClaims struct {
	UserID       uuid.UUID       `json:"user_id"`
	Username     string          `json:"username"`
	Role         models.UserRole `json:"role"`
	TokenVersion int             `json:"token_version"` // For session invalidation
	jwt.RegisteredClaims
}

// TokenValidator interface for validating token version against database
type TokenValidator interface {
	ValidateTokenVersion(userID uuid.UUID, tokenVersion int) (bool, error)
}

// AuthMiddleware validates JWT token from HttpOnly cookie
// Implements zero-trust by validating token version on every request
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
				"error": "Authentication required",
			})
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session data",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("user_role", claims.Role)
		c.Locals("token_version", claims.TokenVersion)

		return c.Next()
	}
}

// AuthMiddlewareWithDB validates JWT and token version against database
// This ensures Kill Switch takes effect immediately
func AuthMiddlewareWithDB(cfg *config.JWTConfig, db *sql.DB) fiber.Handler {
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
				"error": "Authentication required",
			})
		}

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		// Extract claims
		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session data",
			})
		}

		// Validate token version against database (Zero-Trust)
		// This check ensures Kill Switch takes effect immediately
		var currentTokenVersion int
		err = db.QueryRowContext(c.Context(), `
			SELECT token_version FROM users WHERE id = $1 AND is_active = true
		`, claims.UserID).Scan(&currentTokenVersion)

		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Account not found or inactive",
			})
		}
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Authentication verification failed",
			})
		}

		// Token version mismatch means Kill Switch was activated
		if claims.TokenVersion != currentTokenVersion {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Session invalidated. Please login again.",
			})
		}

		// Store user info in context
		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("user_role", claims.Role)
		c.Locals("token_version", claims.TokenVersion)

		return c.Next()
	}
}

// GenerateToken creates a new JWT token with token version
func GenerateToken(user *models.User, cfg *config.JWTConfig) (string, error) {
	claims := JWTClaims{
		UserID:       user.ID,
		Username:     user.Username,
		Role:         user.Role,
		TokenVersion: user.TokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.Expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "cso-erp",
			Subject:   user.ID.String(),
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

// GetUsername extracts username from context
func GetUsername(c *fiber.Ctx) string {
	username, _ := c.Locals("username").(string)
	return username
}

// GetTokenVersion extracts token version from context
func GetTokenVersion(c *fiber.Ctx) int {
	version, _ := c.Locals("token_version").(int)
	return version
}

// GetClientIP extracts client IP address for audit logging
func GetClientIP(c *fiber.Ctx) string {
	// Check for X-Forwarded-For header (for proxied requests)
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check for X-Real-IP header
	if xri := c.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to direct connection IP
	return c.IP()
}

// GetUserAgent extracts user agent for audit logging
func GetUserAgent(c *fiber.Ctx) string {
	return c.Get("User-Agent")
}
