package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TransferHandler struct {
	transferService *services.TransferService
}

func NewTransferHandler(transferService *services.TransferService) *TransferHandler {
	return &TransferHandler{
		transferService: transferService,
	}
}

// CreateTransfer handles transfer creation between containers
func (h *TransferHandler) CreateTransfer(c *fiber.Ctx) error {
	var req models.CreateTransferRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Validate required fields
	if req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Amount must be greater than 0",
		})
	}
	if req.FromContainerID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Source container ID is required",
		})
	}
	if req.ToContainerID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Destination container ID is required",
		})
	}
	if req.FromContainerID == req.ToContainerID {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Source and destination containers must be different",
		})
	}
	if req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description is required",
		})
	}

	userID := middleware.GetUserID(c)

	// Create transfer
	response, err := h.transferService.CreateTransfer(c.Context(), &req, userID)
	if err != nil {
		if err == services.ErrInsufficientBalance {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Insufficient balance in source container",
			})
		}
		if err == services.ErrContainerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Container not found",
			})
		}
		// Check for currency mismatch
		if err.Error() == "currency mismatch" ||
			len(err.Error()) > 17 && err.Error()[:17] == "currency mismatch" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create transfer",
			"details": err.Error(),
		})
	}

	// Success
	return c.Status(fiber.StatusCreated).JSON(response)
}
