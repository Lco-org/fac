package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
)

// DashboardHandler handles dashboard-related requests
// Multi-tenant: creates service per-request from context DB
type DashboardHandler struct{}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	// Note: dashboardService param is kept for backward compatibility but unused
	// Handler now creates service per-request from tenant DB
	return &DashboardHandler{}
}

// getService creates a tenant-scoped DashboardService from request context
func (h *DashboardHandler) getService(c *fiber.Ctx) *services.DashboardService {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil
	}
	return services.NewDashboardService(db)
}

func (h *DashboardHandler) GetWidgetConfigs(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	widgets, err := svc.GetWidgetConfigs(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch widget configurations",
		})
	}

	return c.JSON(widgets)
}

func (h *DashboardHandler) UpdateWidgetConfig(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	widgetID := c.Params("id")

	var req models.UpdateWidgetConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := svc.UpdateWidgetConfig(c.Context(), widgetID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update widget configuration",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Widget configuration updated successfully",
	})
}

func (h *DashboardHandler) BulkUpdateWidgetOrder(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var req []models.WidgetOrderUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := svc.BulkUpdateWidgetOrder(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update widget order",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Widget order updated successfully",
	})
}
