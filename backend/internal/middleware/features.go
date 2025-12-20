package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// RequireFeature middleware checks if the tenant has the required feature enabled
func RequireFeature(feature string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant := GetTenant(c)
		if tenant == nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error": "Tenant context required for feature check",
			})
		}

		hasFeature := false
		for _, f := range tenant.Features {
			if f == feature {
				hasFeature = true
				break
			}
		}

		if !hasFeature {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   "Feature not enabled for this subscription plan",
				"feature": feature,
			})
		}

		return c.Next()
	}
}
