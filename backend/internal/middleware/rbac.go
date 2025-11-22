package middleware

import (
	"cso-erp/internal/models"

	"github.com/gofiber/fiber/v2"
)

// RequireRole creates middleware that enforces role-based access control
func RequireRole(allowedRoles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := GetUserRole(c)

		// Check if user has required role
		for _, role := range allowedRoles {
			if userRole == role {
				return c.Next()
			}
		}

		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error":          "Insufficient permissions",
			"required_roles": allowedRoles,
			"your_role":      userRole,
		})
	}
}

// RequireAdmin middleware - Admin only
func RequireAdmin() fiber.Handler {
	return RequireRole(models.RoleAdmin)
}

// RequireFinanceOrAdmin middleware - Finance or Admin
func RequireFinanceOrAdmin() fiber.Handler {
	return RequireRole(models.RoleAdmin, models.RoleFinance)
}

// RequireAnyAuthenticated - Any authenticated user
func RequireAnyAuthenticated() fiber.Handler {
	return RequireRole(models.RoleAdmin, models.RoleFinance, models.RoleAuditor)
}

// CanDelete checks if user can delete (only Admin can delete, Finance can only reverse)
func CanDelete(c *fiber.Ctx) bool {
	userRole := GetUserRole(c)
	return userRole == models.RoleAdmin
}

// CanWrite checks if user can create/update
func CanWrite(c *fiber.Ctx) bool {
	userRole := GetUserRole(c)
	return userRole == models.RoleAdmin || userRole == models.RoleFinance
}

// CanRead checks if user can read
func CanRead(c *fiber.Ctx) bool {
	userRole := GetUserRole(c)
	return userRole == models.RoleAdmin || userRole == models.RoleFinance || userRole == models.RoleAuditor
}
