package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// EmployeeHandler handles employee-related requests
// Multi-tenant: creates service per-request from context DB
type EmployeeHandler struct{}

func NewEmployeeHandler(employeeService *services.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{}
}

// getService creates a tenant-scoped EmployeeService from request context
func (h *EmployeeHandler) getService(c *fiber.Ctx) *services.EmployeeService {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil
	}
	return services.NewEmployeeService(db)
}

func (h *EmployeeHandler) ListEmployees(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection not found"})
	}

	employees, err := svc.ListEmployees(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch employees"})
	}
	return c.JSON(employees)
}

func (h *EmployeeHandler) GetEmployee(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection not found"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid employee ID"})
	}

	employee, err := svc.GetEmployeeByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
	}
	return c.JSON(employee)
}

func (h *EmployeeHandler) CreateEmployee(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection not found"})
	}

	var req models.CreateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.EmployeeNumber == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Employee number and name are required"})
	}

	userID := middleware.GetUserID(c)
	employee, err := svc.CreateEmployee(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(employee)
}

func (h *EmployeeHandler) UpdateEmployee(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection not found"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid employee ID"})
	}

	var req models.UpdateEmployeeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	employee, err := svc.UpdateEmployee(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(employee)
}

func (h *EmployeeHandler) DeleteEmployee(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Database connection not found"})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid employee ID"})
	}

	if err := svc.DeleteEmployee(c.Context(), id); err != nil {
		if err == services.ErrEmployeeNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Employee not found"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Employee deleted successfully"})
}
