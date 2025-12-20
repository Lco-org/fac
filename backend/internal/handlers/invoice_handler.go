package handlers

import (
	apperrors "cso-erp/internal/errors"
	"cso-erp/internal/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// InvoiceHandler handles invoice requests
type InvoiceHandler struct {
	invoiceService *services.InvoiceService
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService *services.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{invoiceService: invoiceService}
}

// GenerateFromTransaction creates an invoice from a transaction
func (h *InvoiceHandler) GenerateFromTransaction(c *fiber.Ctx) error {
	transactionIDStr := c.Params("transaction_id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(uuid.UUID); ok {
		userID = &uid
	}

	invoice, err := h.invoiceService.GenerateInvoiceFromTransaction(c.Context(), transactionID, userID)
	if err != nil {
		return apperrors.ErrInternal.WithError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Invoice generated successfully",
		"invoice": invoice,
	})
}

// GetInvoice retrieves an invoice by ID
func (h *InvoiceHandler) GetInvoice(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	invoice, err := h.invoiceService.GetInvoice(c.Context(), id)
	if err != nil {
		return apperrors.ErrNotFound
	}

	return c.JSON(fiber.Map{
		"invoice": invoice,
	})
}

// ListInvoices lists all invoices
func (h *InvoiceHandler) ListInvoices(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	if limit > 100 {
		limit = 100
	}

	invoices, err := h.invoiceService.ListInvoices(c.Context(), limit, offset)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"invoices": invoices,
		"limit":    limit,
		"offset":   offset,
	})
}

// PrintInvoice returns HTML version for printing
func (h *InvoiceHandler) PrintInvoice(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	invoice, err := h.invoiceService.GetInvoice(c.Context(), id)
	if err != nil {
		return apperrors.ErrNotFound
	}

	// TODO: Get org info from context/config
	orgName := "منظمة المجتمع المدني"
	orgAddress := "ليبيا - طرابلس"

	html := h.invoiceService.GenerateInvoiceHTML(invoice, orgName, orgAddress)

	c.Set("Content-Type", "text/html; charset=utf-8")
	return c.SendString(html)
}
