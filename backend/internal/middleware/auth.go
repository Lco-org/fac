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

type JWTClaims struct {
	UserID       uuid.UUID       `json:"user_id"`
	Username     string          `json:"username"`
	Role         models.UserRole `json:"role"`
	TokenVersion int             `json:"token_version"`
	jwt.RegisteredClaims
}

type TokenValidator interface {
	ValidateTokenVersion(userID uuid.UUID, tokenVersion int) (bool, error)
}

func AuthMiddleware(cfg *config.JWTConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("auth_token")

		if tokenString == "" {
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

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session data",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("user_role", claims.Role)
		c.Locals("token_version", claims.TokenVersion)

		return c.Next()
	}
}

func AuthMiddlewareWithDB(cfg *config.JWTConfig, db *sql.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Cookies("auth_token")

		if tokenString == "" {
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

		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.Secret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired session",
			})
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid session data",
			})
		}

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

		if claims.TokenVersion != currentTokenVersion {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Session invalidated. Please login again.",
			})
		}

		c.Locals("user_id", claims.UserID)
		c.Locals("username", claims.Username)
		c.Locals("user_role", claims.Role)
		c.Locals("token_version", claims.TokenVersion)

		return c.Next()
	}
}

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

func GetUserID(c *fiber.Ctx) uuid.UUID {
	userID, _ := c.Locals("user_id").(uuid.UUID)
	return userID
}

func GetUserRole(c *fiber.Ctx) models.UserRole {
	role, _ := c.Locals("user_role").(models.UserRole)
	return role
}

func GetUsername(c *fiber.Ctx) string {
	username, _ := c.Locals("username").(string)
	return username
}

func GetTokenVersion(c *fiber.Ctx) int {
	version, _ := c.Locals("token_version").(int)
	return version
}

func GetClientIP(c *fiber.Ctx) string {
	if xff := c.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := c.Get("X-Real-IP"); xri != "" {
		return xri
	}

	return c.IP()
}

func GetUserAgent(c *fiber.Ctx) string {
	return c.Get("User-Agent")
}
