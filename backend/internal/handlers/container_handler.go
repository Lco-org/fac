package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ContainerHandler struct {
	containerService *services.ContainerService
}

func NewContainerHandler(containerService *services.ContainerService) *ContainerHandler {
	return &ContainerHandler{
		containerService: containerService,
	}
}

// ListContainers returns all containers
func (h *ContainerHandler) ListContainers(c *fiber.Ctx) error {
	containers, err := h.containerService.ListContainers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch containers",
		})
	}

	return c.JSON(containers)
}

// GetContainer returns a single container
func (h *ContainerHandler) GetContainer(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid container ID",
		})
	}

	container, err := h.containerService.GetContainerByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Container not found",
		})
	}

	return c.JSON(container)
}

// CreateContainer creates a new container (Admin only)
func (h *ContainerHandler) CreateContainer(c *fiber.Ctx) error {
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

	container, err := h.containerService.CreateContainer(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create container",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(container)
}

// UpdateContainer updates a container (Admin only)
func (h *ContainerHandler) UpdateContainer(c *fiber.Ctx) error {
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

	container, err := h.containerService.UpdateContainer(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update container",
		})
	}

	return c.JSON(container)
}

// DeleteContainer deletes a container (Admin only)
func (h *ContainerHandler) DeleteContainer(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid container ID",
		})
	}

	if err := h.containerService.DeleteContainer(c.Context(), id); err != nil {
		if err == services.ErrContainerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Container not found",
			})
		}
		// Check for specific error messages or types
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
