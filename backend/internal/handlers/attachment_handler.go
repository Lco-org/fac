package handlers

import (
	apperrors "cso-erp/internal/errors"
	"cso-erp/internal/services"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AttachmentHandler handles file attachment requests
type AttachmentHandler struct {
	fileService *services.FileService
}

// NewAttachmentHandler creates a new attachment handler
func NewAttachmentHandler(fileService *services.FileService) *AttachmentHandler {
	return &AttachmentHandler{fileService: fileService}
}

// UploadAttachment uploads a file attachment
func (h *AttachmentHandler) UploadAttachment(c *fiber.Ctx) error {
	entityType := c.FormValue("entity_type")
	entityIDStr := c.FormValue("entity_id")

	if entityType == "" || entityIDStr == "" {
		return apperrors.ErrRequiredField
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	file, err := c.FormFile("file")
	if err != nil {
		return apperrors.New("FILE_002", "File is required", 400)
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(uuid.UUID); ok {
		userID = &uid
	}

	// Get tenant slug from context (set by tenant resolver middleware)
	// This ensures each tenant's files are stored in isolated folders
	tenantSlug := ""
	if slug, ok := c.Locals("tenant_slug").(string); ok {
		tenantSlug = slug
	}
	if tenantSlug == "" {
		// Fallback: try to get from tenant_db
		if tdb, ok := c.Locals("tenant_db").(string); ok && tdb != "" {
			tenantSlug = tdb
		} else {
			tenantSlug = "default"
		}
	}

	attachment, err := h.fileService.UploadAttachment(c.Context(), file, entityType, entityID, userID, tenantSlug)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":    "File uploaded successfully",
		"attachment": attachment,
	})
}

// GetAttachments retrieves attachments for an entity
func (h *AttachmentHandler) GetAttachments(c *fiber.Ctx) error {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		return apperrors.ErrRequiredField
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	attachments, err := h.fileService.GetAttachments(c.Context(), entityType, entityID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"attachments": attachments,
	})
}

// DownloadAttachment downloads an attachment
func (h *AttachmentHandler) DownloadAttachment(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	attachment, err := h.fileService.GetAttachmentByID(c.Context(), id)
	if err != nil {
		return err
	}

	// In a real production app with S3, this would generate a presigned URL
	// For local files, we serve the file (ensure uploadDir is accessible or handle stream)
	// Since fileService stores full relative path, we need base dir.
	// For this MVP, we assume uploads are in "./uploads" relative to CWD
	fullPath := fmt.Sprintf("./uploads/%s", attachment.FilePath)

	return c.Download(fullPath, attachment.FileName)
}

// DeleteAttachment deletes an attachment
func (h *AttachmentHandler) DeleteAttachment(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	if err := h.fileService.DeleteAttachment(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Attachment deleted successfully",
	})
}
