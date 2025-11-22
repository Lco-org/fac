package handlers

import (
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
)

type ReportHandler struct {
	reportService *services.ReportService
}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// GetDailyPosition returns daily financial position report
func (h *ReportHandler) GetDailyPosition(c *fiber.Ctx) error {
	report, err := h.reportService.GetDailyPosition(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate daily position report",
		})
	}

	return c.JSON(report)
}

// GetProjectSummary returns project financial summary
func (h *ReportHandler) GetProjectSummary(c *fiber.Ctx) error {
	projects, err := h.reportService.GetProjectSummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate project summary",
		})
	}

	return c.JSON(projects)
}

// GetDeficitReport returns deficit report
func (h *ReportHandler) GetDeficitReport(c *fiber.Ctx) error {
	report, err := h.reportService.GetDeficitReport(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate deficit report",
		})
	}

	return c.JSON(report)
}
