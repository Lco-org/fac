package handlers

import (
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
)

type EmailHandler struct {
	emailService    *services.EmailService
	userService     *services.UserService
	employeeService *services.EmployeeService
	memberService   *services.MemberService
}

func NewEmailHandler(
	emailService *services.EmailService,
	userService *services.UserService,
	employeeService *services.EmployeeService,
	memberService *services.MemberService,
) *EmailHandler {
	return &EmailHandler{
		emailService:    emailService,
		userService:     userService,
		employeeService: employeeService,
		memberService:   memberService,
	}
}

type BulkEmailRequest struct {
	RecipientType string   `json:"recipient_type"` // "users", "employees", "members"
	RecipientIDs  []string `json:"recipient_ids"`  // UUIDs of selected recipients
	Subject       string   `json:"subject"`
	Message       string   `json:"message"`
}

type BulkEmailResponse struct {
	Total     int      `json:"total"`
	Sent      int      `json:"sent"`
	Failed    int      `json:"failed"`
	FailedIDs []string `json:"failed_ids,omitempty"`
}

func (h *EmailHandler) SendBulkEmail(c *fiber.Ctx) error {
	var req BulkEmailRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Subject == "" || req.Message == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Subject and message are required",
		})
	}

	if len(req.RecipientIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one recipient is required",
		})
	}

	// Get recipients based on type
	var recipients []services.EmailRecipient
	var err error

	switch req.RecipientType {
	case "users":
		recipients, err = h.getUserRecipients(c, req.RecipientIDs)
	case "employees":
		recipients, err = h.getEmployeeRecipients(c, req.RecipientIDs)
	case "members":
		recipients, err = h.getMemberRecipients(c, req.RecipientIDs)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid recipient type",
		})
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch recipients",
		})
	}

	// Send emails
	result := h.emailService.SendBulkEmail(c.Context(), recipients, req.Subject, req.Message)

	return c.JSON(result)
}

func (h *EmailHandler) getUserRecipients(c *fiber.Ctx, ids []string) ([]services.EmailRecipient, error) {
	users, err := h.userService.ListUsers(c.Context())
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var recipients []services.EmailRecipient
	for _, u := range users {
		if idSet[u.ID.String()] {
			name := u.FullName
			if u.FullNameAr != nil && *u.FullNameAr != "" {
				name = *u.FullNameAr
			}
			recipients = append(recipients, services.EmailRecipient{
				ID:    u.ID.String(),
				Email: u.Email,
				Name:  name,
			})
		}
	}
	return recipients, nil
}

func (h *EmailHandler) getEmployeeRecipients(c *fiber.Ctx, ids []string) ([]services.EmailRecipient, error) {
	employees, err := h.employeeService.ListEmployees(c.Context())
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var recipients []services.EmailRecipient
	for _, e := range employees {
		if idSet[e.ID.String()] && e.Email != nil && *e.Email != "" {
			name := e.FullName
			if e.FullNameAr != nil && *e.FullNameAr != "" {
				name = *e.FullNameAr
			}
			recipients = append(recipients, services.EmailRecipient{
				ID:    e.ID.String(),
				Email: *e.Email,
				Name:  name,
			})
		}
	}
	return recipients, nil
}

func (h *EmailHandler) getMemberRecipients(c *fiber.Ctx, ids []string) ([]services.EmailRecipient, error) {
	members, err := h.memberService.ListMembers(c.Context())
	if err != nil {
		return nil, err
	}

	idSet := make(map[string]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var recipients []services.EmailRecipient
	for _, m := range members {
		if idSet[m.ID.String()] && m.Email != nil && *m.Email != "" {
			name := m.FullName
			if m.FullNameAr != nil && *m.FullNameAr != "" {
				name = *m.FullNameAr
			}
			recipients = append(recipients, services.EmailRecipient{
				ID:    m.ID.String(),
				Email: *m.Email,
				Name:  name,
			})
		}
	}
	return recipients, nil
}
