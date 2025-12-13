package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// HRHandler handles HR-related requests
// Multi-tenant: creates service per-request from context DB
type HRHandler struct{}

func NewHRHandler(hrService *services.HRService) *HRHandler {
	return &HRHandler{}
}

// getService creates a tenant-scoped HRService from request context
func (h *HRHandler) getService(c *fiber.Ctx) *services.HRService {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil
	}
	return services.NewHRService(db)
}

// CreateDeduction handles creating a new salary deduction
func (h *HRHandler) CreateDeduction(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var req models.CreateSalaryDeductionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	deduction, err := svc.CreateDeduction(c.Context(), &req, userID)
	if err != nil {
		log.Printf("❌ Failed to create deduction: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(deduction)
}

// ListDeductions returns a list of deductions, optionally filtered by employee
func (h *HRHandler) ListDeductions(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var employeeID *uuid.UUID
	if idStr := c.Query("employee_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid employee ID",
			})
		}
		employeeID = &id
	}

	deductions, err := svc.ListDeductions(c.Context(), employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(deductions)
}

// CreateBonus handles creating a new salary bonus
func (h *HRHandler) CreateBonus(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var req models.CreateSalaryBonusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(uuid.UUID)

	bonus, err := svc.CreateBonus(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(bonus)
}

// ListBonuses returns a list of bonuses, optionally filtered by employee
func (h *HRHandler) ListBonuses(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var employeeID *uuid.UUID
	if idStr := c.Query("employee_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid employee ID",
			})
		}
		employeeID = &id
	}

	bonuses, err := svc.ListBonuses(c.Context(), employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(bonuses)
}

// GetEmployeeFinancials returns the financial profile (salary + active items)
func (h *HRHandler) GetEmployeeFinancialProfile(c *fiber.Ctx) error {
	// Placeholder for future implementation
	return c.JSON(fiber.Map{"message": "Not implemented yet"})
}

// GetMonthlySalaries returns the calculated salaries for a specific month
func (h *HRHandler) GetMonthlySalaries(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	monthStr := c.Query("month")
	var monthYear time.Time
	var err error

	if monthStr == "" {
		monthYear = time.Now()
	} else {
		// Expect YYYY-MM
		monthYear, err = time.Parse("2006-01", monthStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid month format. Use YYYY-MM",
			})
		}
	}

	var employeeID *uuid.UUID
	if idStr := c.Query("employee_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid employee ID",
			})
		}
		employeeID = &id
	}

	summaries, err := svc.GetMonthlySalaries(c.Context(), monthYear, employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(summaries)
}
