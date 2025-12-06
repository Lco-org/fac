package handlers

import (
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboardHandler(dashboardService *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

func (h *DashboardHandler) GetWidgetConfigs(c *fiber.Ctx) error {
	widgets, err := h.dashboardService.GetWidgetConfigs(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch widget configurations",
		})
	}

	return c.JSON(widgets)
}

func (h *DashboardHandler) UpdateWidgetConfig(c *fiber.Ctx) error {
	widgetID := c.Params("id")

	var req models.UpdateWidgetConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.dashboardService.UpdateWidgetConfig(c.Context(), widgetID, req)
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
	var req []models.WidgetOrderUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err := h.dashboardService.BulkUpdateWidgetOrder(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update widget order",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Widget order updated successfully",
	})
}
