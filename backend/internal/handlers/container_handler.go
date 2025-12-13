package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ContainerHandler handles container-related requests
// Multi-tenant: creates service per-request from context DB
type ContainerHandler struct{}

func NewContainerHandler(containerService *services.ContainerService) *ContainerHandler {
	// Note: containerService param is kept for backward compatibility but unused
	// Handler now creates service per-request from tenant DB
	return &ContainerHandler{}
}

// getService creates a tenant-scoped ContainerService from request context
func (h *ContainerHandler) getService(c *fiber.Ctx) *services.ContainerService {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil
	}
	return services.NewContainerService(db)
}

func (h *ContainerHandler) ListContainers(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	containers, err := svc.ListContainers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch containers",
		})
	}

	return c.JSON(containers)
}

func (h *ContainerHandler) GetContainer(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid container ID",
		})
	}

	container, err := svc.GetContainerByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Container not found",
		})
	}

	return c.JSON(container)
}

func (h *ContainerHandler) CreateContainer(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var req models.FinancialContainer
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.Type == "" || req.Currency == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name, type, and currency are required",
		})
	}

	userID := middleware.GetUserID(c)

	container, err := svc.CreateContainer(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create container",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(container)
}

func (h *ContainerHandler) UpdateContainer(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid container ID",
		})
	}

	var req models.FinancialContainer
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	container, err := svc.UpdateContainer(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update container",
		})
	}

	return c.JSON(container)
}

func (h *ContainerHandler) DeleteContainer(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid container ID",
		})
	}

	if err := svc.DeleteContainer(c.Context(), id); err != nil {
		if err == services.ErrContainerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Container not found",
			})
		}
		if err.Error() == "cannot delete container: associated transactions exist (try deactivating instead)" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete container: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Container deleted successfully",
	})
}
