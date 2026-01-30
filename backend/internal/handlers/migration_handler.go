package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
)

// MigrationHandler handles database migration requests for SSA
type MigrationHandler struct {
	migrationService *services.MigrationService
}

func NewMigrationHandler(migrationService *services.MigrationService) *MigrationHandler {
	return &MigrationHandler{
		migrationService: migrationService,
	}
}

// GetMigrationStatus returns the status of migrations for all databases
// GET /api/v1/ssa/migrations/status
func (h *MigrationHandler) GetMigrationStatus(c *fiber.Ctx) error {
	status, err := h.migrationService.GetMigrationStatus(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get migration status",
		})
	}

	return c.JSON(fiber.Map{
		"databases": status,
	})
}

// RunMigrations runs pending migrations on all tenant databases
// POST /api/v1/ssa/migrations/run
func (h *MigrationHandler) RunMigrations(c *fiber.Ctx) error {
	adminID := middleware.GetUserID(c)

	result, err := h.migrationService.RunMigrations(c.Context(), adminID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to run migrations",
			"details": err.Error(),
		})
	}

	statusCode := fiber.StatusOK
	if !result.Success {
		statusCode = fiber.StatusPartialContent
	}

	return c.Status(statusCode).JSON(result)
}
