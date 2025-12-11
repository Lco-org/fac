package handlers

import (
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type HRHandler struct {
	hrService *services.HRService
}

func NewHRHandler(hrService *services.HRService) *HRHandler {
	return &HRHandler{hrService: hrService}
}

// CreateDeduction handles creating a new salary deduction
func (h *HRHandler) CreateDeduction(c *fiber.Ctx) error {
	var req models.CreateSalaryDeductionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("userID").(uuid.UUID)

	deduction, err := h.hrService.CreateDeduction(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(deduction)
}

// ListDeductions returns a list of deductions, optionally filtered by employee
func (h *HRHandler) ListDeductions(c *fiber.Ctx) error {
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

	deductions, err := h.hrService.ListDeductions(c.Context(), employeeID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(deductions)
}

// CreateBonus handles creating a new salary bonus
func (h *HRHandler) CreateBonus(c *fiber.Ctx) error {
	var req models.CreateSalaryBonusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("userID").(uuid.UUID)

	bonus, err := h.hrService.CreateBonus(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(bonus)
}

// ListBonuses returns a list of bonuses, optionally filtered by employee
func (h *HRHandler) ListBonuses(c *fiber.Ctx) error {
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

	bonuses, err := h.hrService.ListBonuses(c.Context(), employeeID)
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
