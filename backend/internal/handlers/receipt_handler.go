package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ReceiptHandler handles receipt-related requests
// Multi-tenant: creates service per-request from context DB
type ReceiptHandler struct{}

func NewReceiptHandler(receiptService *services.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{}
}

// getService creates a tenant-scoped ReceiptService from request context
func (h *ReceiptHandler) getService(c *fiber.Ctx) *services.ReceiptService {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil
	}
	return services.NewReceiptService(db)
}

func (h *ReceiptHandler) GetReceiptData(c *fiber.Ctx) error {
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
			"error": "Invalid transaction ID",
		})
	}

	receiptData, err := svc.GetReceiptData(c.Context(), id)
	if err != nil {
		if err.Error() == "transaction not found" {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Transaction not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve receipt data",
		})
	}

	return c.JSON(receiptData)
}
