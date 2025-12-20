package middleware

import (
	"cso-erp/internal/models"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// ===============================================
// RBAC MIDDLEWARE TESTS
// ===============================================

func TestRequireRole_Allowed(t *testing.T) {
	app := fiber.New()

	// Set user role in context before RBAC check
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", models.RoleAdmin)
		return c.Next()
	})
	app.Use(RequireRole(models.RoleAdmin, models.RoleFinance))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRequireRole_Denied(t *testing.T) {
	app := fiber.New()

	// Set user role to Auditor
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_role", models.RoleAuditor)
		return c.Next()
	})
	// Require Admin or Finance only
	app.Use(RequireRole(models.RoleAdmin, models.RoleFinance))
	app.Get("/protected", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	resp, err := app.Test(req)

	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.StatusCode)
	}
}

func TestRequireAdmin(t *testing.T) {
	tests := []struct {
		name           string
		role           models.UserRole
		expectedStatus int
	}{
		{"Admin allowed", models.RoleAdmin, fiber.StatusOK},
		{"Finance denied", models.RoleFinance, fiber.StatusForbidden},
		{"Auditor denied", models.RoleAuditor, fiber.StatusForbidden},
		{"HR denied", models.RoleHR, fiber.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				return c.Next()
			})
			app.Use(RequireAdmin())
			app.Get("/admin", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/admin", nil)
			resp, _ := app.Test(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRequireFinanceOrAdmin(t *testing.T) {
	tests := []struct {
		name           string
		role           models.UserRole
		expectedStatus int
	}{
		{"Admin allowed", models.RoleAdmin, fiber.StatusOK},
		{"Finance allowed", models.RoleFinance, fiber.StatusOK},
		{"Auditor denied", models.RoleAuditor, fiber.StatusForbidden},
		{"HR denied", models.RoleHR, fiber.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				return c.Next()
			})
			app.Use(RequireFinanceOrAdmin())
			app.Get("/finance", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/finance", nil)
			resp, _ := app.Test(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

func TestRequireAnyAuthenticated(t *testing.T) {
	tests := []struct {
		name           string
		role           models.UserRole
		expectedStatus int
	}{
		{"Admin allowed", models.RoleAdmin, fiber.StatusOK},
		{"Finance allowed", models.RoleFinance, fiber.StatusOK},
		{"Auditor allowed", models.RoleAuditor, fiber.StatusOK},
		{"HR allowed", models.RoleHR, fiber.StatusOK},
		{"No role denied", "", fiber.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			app.Use(func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				return c.Next()
			})
			app.Use(RequireAnyAuthenticated())
			app.Get("/any", func(c *fiber.Ctx) error {
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/any", nil)
			resp, _ := app.Test(req)

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}
		})
	}
}

// ===============================================
// PERMISSION HELPER TESTS
// ===============================================

func TestCanDelete(t *testing.T) {
	tests := []struct {
		name     string
		role     models.UserRole
		expected bool
	}{
		{"Admin can delete", models.RoleAdmin, true},
		{"Finance cannot delete", models.RoleFinance, false},
		{"Auditor cannot delete", models.RoleAuditor, false},
		{"HR cannot delete", models.RoleHR, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var result bool
			app.Get("/test", func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				result = CanDelete(c)
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			app.Test(req)

			if result != tt.expected {
				t.Errorf("CanDelete() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCanWrite(t *testing.T) {
	tests := []struct {
		name     string
		role     models.UserRole
		expected bool
	}{
		{"Admin can write", models.RoleAdmin, true},
		{"Finance can write", models.RoleFinance, true},
		{"HR can write", models.RoleHR, true},
		{"Auditor cannot write", models.RoleAuditor, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var result bool
			app.Get("/test", func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				result = CanWrite(c)
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			app.Test(req)

			if result != tt.expected {
				t.Errorf("CanWrite() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestCanRead(t *testing.T) {
	tests := []struct {
		name     string
		role     models.UserRole
		expected bool
	}{
		{"Admin can read", models.RoleAdmin, true},
		{"Finance can read", models.RoleFinance, true},
		{"Auditor can read", models.RoleAuditor, true},
		{"HR can read", models.RoleHR, true},
		{"No role cannot read", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()

			var result bool
			app.Get("/test", func(c *fiber.Ctx) error {
				c.Locals("user_role", tt.role)
				result = CanRead(c)
				return c.SendString("OK")
			})

			req := httptest.NewRequest("GET", "/test", nil)
			app.Test(req)

			if result != tt.expected {
				t.Errorf("CanRead() = %v, want %v", result, tt.expected)
			}
		})
	}
}
