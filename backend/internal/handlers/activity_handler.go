package handlers

import (
	apperrors "cso-erp/internal/errors"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ActivityHandler handles activity log requests
type ActivityHandler struct {
	activityService *services.ActivityService
}

// NewActivityHandler creates a new activity handler
func NewActivityHandler(activityService *services.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

// GetEntityActivities retrieves activities for an entity
func (h *ActivityHandler) GetEntityActivities(c *fiber.Ctx) error {
	entityType := c.Query("entity_type")
	entityIDStr := c.Query("entity_id")

	if entityType == "" || entityIDStr == "" {
		return apperrors.ErrRequiredField
	}

	entityID, err := uuid.Parse(entityIDStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	activities, err := h.activityService.GetEntityActivities(c.Context(), entityType, entityID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"activities": activities,
	})
}

// GetUserActivities retrieves activities for a user
func (h *ActivityHandler) GetUserActivities(c *fiber.Ctx) error {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		// Use current user if not specified
		if uid, ok := c.Locals("user_id").(uuid.UUID); ok {
			userIDStr = uid.String()
		} else {
			return apperrors.ErrRequiredField
		}
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	activities, err := h.activityService.GetUserActivities(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"activities": activities,
	})
}
