package handlers

import (
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/services"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AssetHandler handles asset-related requests
type AssetHandler struct{}

func NewAssetHandler() *AssetHandler {
	return &AssetHandler{}
}

// getService creates a tenant-scoped AssetService from request context
func (h *AssetHandler) getService(c *fiber.Ctx) *services.AssetService {
	db := middleware.GetTenantDB(c)
	if db == nil {
		return nil
	}
	return services.NewAssetService(db)
}

// ==================== CATEGORIES ====================

// ListCategories returns all asset categories
// GET /api/v1/asset-categories
func (h *AssetHandler) ListCategories(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	categories, err := svc.ListCategories(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list categories",
		})
	}

	return c.JSON(fiber.Map{
		"categories": categories,
	})
}

// CreateCategory creates a new asset category
// POST /api/v1/asset-categories
func (h *AssetHandler) CreateCategory(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var req models.CreateAssetCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" || req.NameAr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Name and name_ar are required",
		})
	}

	category, err := svc.CreateCategory(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create category",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(category)
}

// UpdateCategory updates an asset category
// PUT /api/v1/asset-categories/:id
func (h *AssetHandler) UpdateCategory(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	var req models.CreateAssetCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := svc.UpdateCategory(c.Context(), id, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update category",
		})
	}

	return c.JSON(fiber.Map{"message": "Category updated successfully"})
}

// DeleteCategory deletes an asset category
// DELETE /api/v1/asset-categories/:id
func (h *AssetHandler) DeleteCategory(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category ID",
		})
	}

	if err := svc.DeleteCategory(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete category",
		})
	}

	return c.JSON(fiber.Map{"message": "Category deleted successfully"})
}

// ==================== TYPES ====================

// ListTypes returns all asset types
// GET /api/v1/asset-types
func (h *AssetHandler) ListTypes(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var categoryID *uuid.UUID
	if catIDStr := c.Query("category_id"); catIDStr != "" {
		if catID, err := uuid.Parse(catIDStr); err == nil {
			categoryID = &catID
		}
	}

	types, err := svc.ListTypes(c.Context(), categoryID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list types",
		})
	}

	return c.JSON(fiber.Map{
		"types": types,
	})
}

// CreateType creates a new asset type
// POST /api/v1/asset-types
func (h *AssetHandler) CreateType(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var req models.CreateAssetTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.CategoryID == uuid.Nil || req.Name == "" || req.NameAr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "category_id, name, and name_ar are required",
		})
	}

	assetType, err := svc.CreateType(c.Context(), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create type",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(assetType)
}

// DeleteType deletes an asset type
// DELETE /api/v1/asset-types/:id
func (h *AssetHandler) DeleteType(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid type ID",
		})
	}

	if err := svc.DeleteType(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete type",
		})
	}

	return c.JSON(fiber.Map{"message": "Type deleted successfully"})
}

// ==================== ASSETS ====================

// ListAssets returns assets with optional filters
// GET /api/v1/assets
func (h *AssetHandler) ListAssets(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	filter := &models.AssetListFilter{
		Status:          c.Query("status"),
		Condition:       c.Query("condition"),
		AcquisitionType: c.Query("acquisition_type"),
		Search:          c.Query("search"),
		Page:            c.QueryInt("page", 1),
		Limit:           c.QueryInt("limit", 50),
	}

	if catID := c.Query("category_id"); catID != "" {
		if id, err := uuid.Parse(catID); err == nil {
			filter.CategoryID = &id
		}
	}
	if typeID := c.Query("type_id"); typeID != "" {
		if id, err := uuid.Parse(typeID); err == nil {
			filter.TypeID = &id
		}
	}
	if projID := c.Query("project_id"); projID != "" {
		if id, err := uuid.Parse(projID); err == nil {
			filter.ProjectID = &id
		}
	}

	assets, total, err := svc.ListAssets(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list assets",
		})
	}

	return c.JSON(fiber.Map{
		"assets": assets,
		"total":  total,
		"page":   filter.Page,
		"limit":  filter.Limit,
	})
}

// GetAsset returns a single asset by ID
// GET /api/v1/assets/:id
func (h *AssetHandler) GetAsset(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid asset ID",
		})
	}

	asset, err := svc.GetAsset(c.Context(), id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Asset not found",
		})
	}

	return c.JSON(asset)
}

// CreateAsset creates a new asset
// POST /api/v1/assets
func (h *AssetHandler) CreateAsset(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	var req models.CreateAssetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if req.AssetTypeID == uuid.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "asset_type_id is required",
		})
	}
	if req.AcquisitionType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "acquisition_type is required",
		})
	}

	userID := middleware.GetUserID(c)

	asset, err := svc.CreateAsset(c.Context(), &req, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to create asset",
			"details": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(asset)
}

// UpdateAsset updates an existing asset
// PUT /api/v1/assets/:id
func (h *AssetHandler) UpdateAsset(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid asset ID",
		})
	}

	var req models.UpdateAssetRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if err := svc.UpdateAsset(c.Context(), id, &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update asset",
		})
	}

	return c.JSON(fiber.Map{"message": "Asset updated successfully"})
}

// DeleteAsset deletes an asset
// DELETE /api/v1/assets/:id
func (h *AssetHandler) DeleteAsset(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid asset ID",
		})
	}

	if err := svc.DeleteAsset(c.Context(), id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete asset",
		})
	}

	return c.JSON(fiber.Map{"message": "Asset deleted successfully"})
}

// GetInventorySummary returns a summary of assets by category
// GET /api/v1/assets/inventory
func (h *AssetHandler) GetInventorySummary(c *fiber.Ctx) error {
	svc := h.getService(c)
	if svc == nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Database connection not found",
		})
	}

	summary, err := svc.GetInventorySummary(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get inventory summary",
		})
	}

	return c.JSON(fiber.Map{
		"summary": summary,
	})
}
