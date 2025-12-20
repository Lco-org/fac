package handlers

import (
	apperrors "cso-erp/internal/errors"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// BackupHandler handles backup requests
type BackupHandler struct {
	backupService *services.BackupService
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(backupService *services.BackupService) *BackupHandler {
	return &BackupHandler{backupService: backupService}
}

// CreateBackup initiates a new backup
func (h *BackupHandler) CreateBackup(c *fiber.Ctx) error {
	// TODO: Get actual DB name from config or context
	dbName := "cso-erp" // Default for now

	var userID *uuid.UUID
	if uid, ok := c.Locals("user_id").(uuid.UUID); ok {
		userID = &uid
	}

	backup, err := h.backupService.CreateBackup(c.Context(), dbName, userID)
	if err != nil {
		return apperrors.ErrInternal.WithError(err)
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Backup created successfully",
		"backup":  backup,
	})
}

// ListBackups lists all backups
func (h *BackupHandler) ListBackups(c *fiber.Ctx) error {
	backups, err := h.backupService.ListBackups(c.Context())
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"backups": backups,
	})
}
