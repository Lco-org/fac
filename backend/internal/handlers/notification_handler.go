package handlers

import (
	apperrors "cso-erp/internal/errors"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// NotificationHandler handles notification requests
type NotificationHandler struct {
	notificationService *services.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService *services.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationService: notificationService}
}

// GetNotifications retrieves notifications for the current user
func (h *NotificationHandler) GetNotifications(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return apperrors.ErrInvalidToken
	}

	notifications, err := h.notificationService.GetUserNotifications(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"notifications": notifications,
	})
}

// GetUnreadCount returns count of unread notifications
func (h *NotificationHandler) GetUnreadCount(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return apperrors.ErrInvalidToken
	}

	count, err := h.notificationService.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"unread_count": count,
	})
}

// MarkAsRead marks a notification as read
func (h *NotificationHandler) MarkAsRead(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return apperrors.ErrInvalidUUID
	}

	if err := h.notificationService.MarkAsRead(c.Context(), id); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Notification marked as read",
	})
}

// MarkAllAsRead marks all notifications as read
func (h *NotificationHandler) MarkAllAsRead(c *fiber.Ctx) error {
	userID, ok := c.Locals("user_id").(uuid.UUID)
	if !ok {
		return apperrors.ErrInvalidToken
	}

	if err := h.notificationService.MarkAllAsRead(c.Context(), userID); err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"message": "All notifications marked as read",
	})
}
