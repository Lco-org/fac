package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ProjectHandler struct {
	projectService *services.ProjectService
}

func NewProjectHandler(projectService *services.ProjectService) *ProjectHandler {
	return &ProjectHandler{
		projectService: projectService,
	}
}

// ListProjects returns all projects
func (h *ProjectHandler) ListProjects(c *fiber.Ctx) error {
	projects, err := h.projectService.ListProjects(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch projects",
		})
	}

	return c.JSON(projects)
}

// GetProject returns a single project
func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	project, err := h.projectService.GetProjectByID(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Project not found",
		})
	}

	return c.JSON(project)
}

// CreateProject creates a new project
func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	log.Printf("DEBUG: CreateProject called")
	log.Printf("DEBUG: Request body: %s", string(c.Body()))

	var req models.CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("ERROR: Failed to parse request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	log.Printf("DEBUG: Parsed request: %+v", req)

	if req.Code == "" || req.Name == "" {
		log.Printf("ERROR: Code or Name is empty - Code: '%s', Name: '%s'", req.Code, req.Name)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Code and name are required",
		})
	}

	userID := middleware.GetUserID(c)
	log.Printf("DEBUG: Creating project for user: %s", userID)

	project, err := h.projectService.CreateProject(c.Context(), &req, userID)
	if err != nil {
		log.Printf("ERROR: Service failed to create project: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	log.Printf("DEBUG: Project created successfully: %+v", project)
	return c.Status(fiber.StatusCreated).JSON(project)
}

// UpdateProject updates a project
func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	var req models.UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	project, err := h.projectService.UpdateProject(c.Context(), id, &req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(project)
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid project ID",
		})
	}

	if err := h.projectService.DeleteProject(c.Context(), id); err != nil {
		if err == services.ErrProjectNotFound {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Project not found",
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Project deleted successfully",
	})
}
