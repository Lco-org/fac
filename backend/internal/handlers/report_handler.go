package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ReportHandler handles report-related requests
// Multi-tenant: creates service per-request from context DB
type ReportHandler struct{}

func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	// Note: reportService param is kept for backward compatibility but unused
	// Handler now creates service per-request from tenant DB
	return &ReportHandler{}
}

// getService creates a tenant-scoped ReportService from request context
func (h *ReportHandler) getService(c *fiber.Ctx) *services.ReportService {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil
	}
	return services.NewReportService(db)
}

// GetDashboardSummary returns dashboard summary
func (h *ReportHandler) GetDashboardSummary(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	// Get user role from context (set by auth middleware)
	userRole := c.Locals("user_role").(models.UserRole)

	summary, err := svc.GetDashboardSummary(c.Context(), string(userRole))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to generate dashboard summary",
			"details": err.Error(),
		})
	}

	return c.JSON(summary)
}

// GetContainerStatement returns detailed financial statement for a container
func (h *ReportHandler) GetContainerStatement(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	containerID := c.Params("id")
	if containerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Container ID is required",
		})
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	if start := c.Query("start_date"); start != "" {
		t, err := time.Parse("2006-01-02", start)
		if err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		t, err := time.Parse("2006-01-02", end)
		if err == nil {
			// Set to end of day
			t = t.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
			endDate = &t
		}
	}

	statement, err := svc.GetContainerStatement(c.Context(), containerID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate container statement",
		})
	}

	return c.JSON(statement)
}

// GetProjectStatement returns detailed financial statement for a project
func (h *ReportHandler) GetProjectStatement(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	projectID := c.Params("id")
	if projectID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Project ID is required",
		})
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	if start := c.Query("start_date"); start != "" {
		t, err := time.Parse("2006-01-02", start)
		if err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		t, err := time.Parse("2006-01-02", end)
		if err == nil {
			// Set to end of day
			t = t.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
			endDate = &t
		}
	}

	statement, err := svc.GetProjectStatement(c.Context(), projectID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate project statement",
		})
	}

	return c.JSON(statement)
}

// GetDailyPosition returns financial containers grouped by type
func (h *ReportHandler) GetDailyPosition(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	position, err := svc.GetDailyPosition(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to generate daily position",
		})
	}

	return c.JSON(position)
}

// GetProjectSummary returns a list of all active projects
func (h *ReportHandler) GetProjectSummary(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	projects, err := svc.GetProjectSummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	return c.JSON(projects)
}

// GetInvoices returns a list of income transactions (invoices)
func (h *ReportHandler) GetInvoices(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	if start := c.Query("start_date"); start != "" {
		t, err := time.Parse("2006-01-02", start)
		if err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		t, err := time.Parse("2006-01-02", end)
		if err == nil {
			// Set to end of day
			t = t.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
			endDate = &t
		}
	}

	projectID := c.Query("project_id")
	containerID := c.Query("container_id")
	transactionNumber := c.Query("transaction_number")

	invoices, err := svc.GetInvoices(c.Context(), startDate, endDate, projectID, containerID, transactionNumber)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch invoices",
		})
	}

	return c.JSON(invoices)
}

// GetEmployeeStatement returns detailed financial statement for an employee
func (h *ReportHandler) GetEmployeeStatement(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	employeeID := c.Params("id")
	if employeeID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Employee ID is required",
		})
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	if start := c.Query("start_date"); start != "" {
		t, err := time.Parse("2006-01-02", start)
		if err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		t, err := time.Parse("2006-01-02", end)
		if err == nil {
			// Set to end of day
			t = t.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
			endDate = &t
		}
	}

	statement, err := svc.GetEmployeeStatement(c.Context(), employeeID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to generate employee statement",
			"details": err.Error(),
		})
	}

	return c.JSON(statement)
}

// GetMemberStatement returns detailed financial statement for a member
func (h *ReportHandler) GetMemberStatement(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	memberID := c.Params("id")
	if memberID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Member ID is required",
		})
	}

	// Parse optional date filters
	var startDate, endDate *time.Time
	if start := c.Query("start_date"); start != "" {
		t, err := time.Parse("2006-01-02", start)
		if err == nil {
			startDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		t, err := time.Parse("2006-01-02", end)
		if err == nil {
			// Set to end of day
			t = t.Add(24 * time.Hour).Add(-1 * time.Nanosecond)
			endDate = &t
		}
	}

	statement, err := svc.GetMemberStatement(c.Context(), memberID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to generate member statement",
			"details": err.Error(),
		})
	}

	return c.JSON(statement)
}
