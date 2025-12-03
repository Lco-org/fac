package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	transactionService *services.TransactionService
}

func NewTransactionHandler(transactionService *services.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// CreateTransaction handles transaction creation with deficit warning
func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	var req models.CreateTransactionRequest
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
	if req.ContainerID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Container ID is required",
		})
	}
	if req.ProjectID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Project ID is required",
		})
	}
	if req.Description == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Description is required",
		})
	}

	// Set default exchange rate if not provided
	if req.ExchangeRate == 0 {
		req.ExchangeRate = 1.0
	}

	userID := middleware.GetUserID(c)

	// Create transaction
	response, err := h.transactionService.CreateTransaction(c.Context(), &req, userID)
	if err != nil {
		if err == services.ErrInsufficientBalance {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Insufficient balance",
			})
		}
		if err == services.ErrContainerNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Container not found",
			})
		}
		if err == services.ErrProjectNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create transaction",
		})
	}

	// If deficit warning, return 202 Accepted with warning
	if response.DeficitWarning && response.RequiresApproval {
		return c.Status(fiber.StatusAccepted).JSON(response)
	}

	// Success
	return c.Status(fiber.StatusCreated).JSON(response)
}

// GetTransaction retrieves a transaction by ID
func (h *TransactionHandler) GetTransaction(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction ID",
		})
	}

	transaction, err := h.transactionService.GetTransactionByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	return c.JSON(transaction)
}

// ReverseTransaction reverses a transaction (Finance/Admin only)
func (h *TransactionHandler) ReverseTransaction(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction ID",
		})
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := c.BodyParser(&req); err != nil || req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Reversal reason is required",
		})
	}

	userID := middleware.GetUserID(c)

	err = h.transactionService.ReverseTransaction(c.Context(), id, req.Reason, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to reverse transaction",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Transaction reversed successfully",
	})
}
