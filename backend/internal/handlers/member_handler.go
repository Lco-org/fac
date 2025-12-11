package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MemberHandler struct {
	memberService *services.MemberService
}

func NewMemberHandler(memberService *services.MemberService) *MemberHandler {
	return &MemberHandler{memberService: memberService}
}

func (h *MemberHandler) ListMembers(c *fiber.Ctx) error {
	members, err := h.memberService.ListMembers(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch members"})
	}
	return c.JSON(members)
}

func (h *MemberHandler) GetMember(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid member ID"})
	}

	member, err := h.memberService.GetMemberByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Member not found"})
	}
	return c.JSON(member)
}

func (h *MemberHandler) CreateMember(c *fiber.Ctx) error {
	var req models.CreateMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.MembershipNumber == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Membership number and name are required"})
	}

	userID := middleware.GetUserID(c)
	member, err := h.memberService.CreateMember(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(member)
}

func (h *MemberHandler) UpdateMember(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid member ID"})
	}

	var req models.UpdateMemberRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	member, err := h.memberService.UpdateMember(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(member)
}

func (h *MemberHandler) DeleteMember(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid member ID"})
	}

	if err := h.memberService.DeleteMember(c.Context(), id); err != nil {
		if err == services.ErrMemberNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Member not found"})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "Member deleted successfully"})
}

// RenewMembership processes a membership renewal/subscription payment
func (h *MemberHandler) RenewMembership(c *fiber.Ctx) error {
	var req models.RenewMembershipRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	if req.MemberID == uuid.Nil || req.Amount <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Member ID and amount are required"})
	}

	userID := middleware.GetUserID(c)
	subscription, err := h.memberService.RenewMembership(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusCreated).JSON(subscription)
}

// GetExpiringMembers returns members whose membership expires within 30 days
func (h *MemberHandler) GetExpiringMembers(c *fiber.Ctx) error {
	// Default to 30 days threshold
	daysThreshold := 30

	members, err := h.memberService.GetExpiringMembers(c.Context(), daysThreshold)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch expiring members"})
	}

	return c.JSON(members)
}

// GetSubscriptions returns subscription history for a member
func (h *MemberHandler) GetSubscriptions(c *fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid member ID"})
	}

	subscriptions, err := h.memberService.GetMemberSubscriptions(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch subscriptions"})
	}

	return c.JSON(subscriptions)
}
