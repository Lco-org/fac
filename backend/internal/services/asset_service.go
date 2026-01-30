package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AssetService handles asset management operations
type AssetService struct {
	db *database.DB
}

func NewAssetService(db *database.DB) *AssetService {
	return &AssetService{db: db}
}

// ==================== CATEGORIES ====================

// ListCategories returns all asset categories
func (s *AssetService) ListCategories(ctx context.Context) ([]models.AssetCategory, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, icon, color, sort_order, is_active, created_at, updated_at
		FROM asset_categories
		WHERE is_active = true
		ORDER BY sort_order ASC, name_ar ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []models.AssetCategory
	for rows.Next() {
		var c models.AssetCategory
		if err := rows.Scan(&c.ID, &c.Name, &c.NameAr, &c.Icon, &c.Color, &c.SortOrder, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// CreateCategory creates a new asset category
func (s *AssetService) CreateCategory(ctx context.Context, req *models.CreateAssetCategoryRequest) (*models.AssetCategory, error) {
	category := &models.AssetCategory{
		ID:        uuid.New(),
		Name:      req.Name,
		NameAr:    req.NameAr,
		Icon:      req.Icon,
		Color:     req.Color,
		SortOrder: req.SortOrder,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if category.Icon == "" {
		category.Icon = "box"
	}
	if category.Color == "" {
		category.Color = "#6366f1"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO asset_categories (id, name, name_ar, icon, color, sort_order, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, category.ID, category.Name, category.NameAr, category.Icon, category.Color, category.SortOrder, category.IsActive, category.CreatedAt, category.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return category, nil
}

// UpdateCategory updates an asset category
func (s *AssetService) UpdateCategory(ctx context.Context, id uuid.UUID, req *models.CreateAssetCategoryRequest) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE asset_categories
		SET name = $1, name_ar = $2, icon = $3, color = $4, sort_order = $5, updated_at = $6
		WHERE id = $7
	`, req.Name, req.NameAr, req.Icon, req.Color, req.SortOrder, time.Now(), id)
	return err
}

// DeleteCategory soft deletes an asset category
func (s *AssetService) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE asset_categories SET is_active = false, updated_at = $1 WHERE id = $2
	`, time.Now(), id)
	return err
}

// ==================== TYPES ====================

// ListTypes returns all asset types, optionally filtered by category
func (s *AssetService) ListTypes(ctx context.Context, categoryID *uuid.UUID) ([]models.AssetType, error) {
	query := `
		SELECT t.id, t.category_id, t.name, t.name_ar, t.unit, t.depreciation_years, t.is_active, t.created_at,
		       c.id, c.name, c.name_ar, c.icon, c.color
		FROM asset_types t
		JOIN asset_categories c ON t.category_id = c.id
		WHERE t.is_active = true
	`
	args := []interface{}{}

	if categoryID != nil {
		query += " AND t.category_id = $1"
		args = append(args, *categoryID)
	}

	query += " ORDER BY c.sort_order ASC, t.name_ar ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list types: %w", err)
	}
	defer rows.Close()

	var types []models.AssetType
	for rows.Next() {
		var t models.AssetType
		var c models.AssetCategory
		if err := rows.Scan(&t.ID, &t.CategoryID, &t.Name, &t.NameAr, &t.Unit, &t.DepreciationYears, &t.IsActive, &t.CreatedAt,
			&c.ID, &c.Name, &c.NameAr, &c.Icon, &c.Color); err != nil {
			return nil, fmt.Errorf("failed to scan type: %w", err)
		}
		t.Category = &c
		types = append(types, t)
	}
	return types, nil
}

// CreateType creates a new asset type
func (s *AssetService) CreateType(ctx context.Context, req *models.CreateAssetTypeRequest) (*models.AssetType, error) {
	assetType := &models.AssetType{
		ID:                uuid.New(),
		CategoryID:        req.CategoryID,
		Name:              req.Name,
		NameAr:            req.NameAr,
		Unit:              req.Unit,
		DepreciationYears: req.DepreciationYears,
		IsActive:          true,
		CreatedAt:         time.Now(),
	}

	if assetType.Unit == "" {
		assetType.Unit = "قطعة"
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO asset_types (id, category_id, name, name_ar, unit, depreciation_years, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, assetType.ID, assetType.CategoryID, assetType.Name, assetType.NameAr, assetType.Unit, assetType.DepreciationYears, assetType.IsActive, assetType.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create type: %w", err)
	}
	return assetType, nil
}

// DeleteType soft deletes an asset type
func (s *AssetService) DeleteType(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE asset_types SET is_active = false WHERE id = $1
	`, id)
	return err
}

// ==================== ASSETS ====================

// ListAssets returns assets with optional filters
func (s *AssetService) ListAssets(ctx context.Context, filter *models.AssetListFilter) ([]models.Asset, int, error) {
	// Build WHERE clause
	whereClause := "WHERE a.is_active = true"
	args := []interface{}{}
	argNum := 1

	if filter.CategoryID != nil {
		whereClause += fmt.Sprintf(" AND t.category_id = $%d", argNum)
		args = append(args, *filter.CategoryID)
		argNum++
	}
	if filter.TypeID != nil {
		whereClause += fmt.Sprintf(" AND a.asset_type_id = $%d", argNum)
		args = append(args, *filter.TypeID)
		argNum++
	}
	if filter.ProjectID != nil {
		whereClause += fmt.Sprintf(" AND a.project_id = $%d", argNum)
		args = append(args, *filter.ProjectID)
		argNum++
	}
	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND a.status = $%d", argNum)
		args = append(args, filter.Status)
		argNum++
	}
	if filter.Condition != "" {
		whereClause += fmt.Sprintf(" AND a.condition = $%d", argNum)
		args = append(args, filter.Condition)
		argNum++
	}
	if filter.AcquisitionType != "" {
		whereClause += fmt.Sprintf(" AND a.acquisition_type = $%d", argNum)
		args = append(args, filter.AcquisitionType)
		argNum++
	}
	if filter.Search != "" {
		whereClause += fmt.Sprintf(" AND (a.name_ar ILIKE $%d OR a.serial_number ILIKE $%d OR a.asset_code ILIKE $%d)", argNum, argNum, argNum)
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM assets a
		JOIN asset_types t ON a.asset_type_id = t.id
		%s
	`, whereClause)

	var total int
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	// Pagination
	if filter.Limit == 0 {
		filter.Limit = 50
	}
	if filter.Page == 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.Limit

	query := fmt.Sprintf(`
		SELECT a.id, a.asset_type_id, a.project_id, a.transaction_id,
		       a.asset_code, a.serial_number, a.name, a.name_ar, a.description,
		       a.quantity, a.unit_price, a.total_value, a.currency,
		       a.condition, a.status, a.acquisition_date, a.acquisition_type,
		       a.supplier, a.location, a.assigned_to, a.notes,
		       a.is_active, a.created_at, a.updated_at, a.created_by,
		       t.id, t.name, t.name_ar, t.unit,
		       c.id, c.name, c.name_ar, c.icon, c.color
		FROM assets a
		JOIN asset_types t ON a.asset_type_id = t.id
		JOIN asset_categories c ON t.category_id = c.id
		%s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, filter.Limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list assets: %w", err)
	}
	defer rows.Close()

	var assets []models.Asset
	for rows.Next() {
		var a models.Asset
		var t models.AssetType
		var c models.AssetCategory

		if err := rows.Scan(
			&a.ID, &a.AssetTypeID, &a.ProjectID, &a.TransactionID,
			&a.AssetCode, &a.SerialNumber, &a.Name, &a.NameAr, &a.Description,
			&a.Quantity, &a.UnitPrice, &a.TotalValue, &a.Currency,
			&a.Condition, &a.Status, &a.AcquisitionDate, &a.AcquisitionType,
			&a.Supplier, &a.Location, &a.AssignedTo, &a.Notes,
			&a.IsActive, &a.CreatedAt, &a.UpdatedAt, &a.CreatedBy,
			&t.ID, &t.Name, &t.NameAr, &t.Unit,
			&c.ID, &c.Name, &c.NameAr, &c.Icon, &c.Color,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan asset: %w", err)
		}

		t.Category = &c
		a.AssetType = &t
		assets = append(assets, a)
	}

	return assets, total, nil
}

// GetAsset returns a single asset by ID
func (s *AssetService) GetAsset(ctx context.Context, id uuid.UUID) (*models.Asset, error) {
	var a models.Asset
	var t models.AssetType
	var c models.AssetCategory

	err := s.db.QueryRowContext(ctx, `
		SELECT a.id, a.asset_type_id, a.project_id, a.transaction_id,
		       a.asset_code, a.serial_number, a.name, a.name_ar, a.description,
		       a.quantity, a.unit_price, a.total_value, a.currency,
		       a.condition, a.status, a.acquisition_date, a.acquisition_type,
		       a.supplier, a.location, a.assigned_to, a.notes,
		       a.is_active, a.created_at, a.updated_at, a.created_by,
		       t.id, t.name, t.name_ar, t.unit,
		       c.id, c.name, c.name_ar, c.icon, c.color
		FROM assets a
		JOIN asset_types t ON a.asset_type_id = t.id
		JOIN asset_categories c ON t.category_id = c.id
		WHERE a.id = $1
	`, id).Scan(
		&a.ID, &a.AssetTypeID, &a.ProjectID, &a.TransactionID,
		&a.AssetCode, &a.SerialNumber, &a.Name, &a.NameAr, &a.Description,
		&a.Quantity, &a.UnitPrice, &a.TotalValue, &a.Currency,
		&a.Condition, &a.Status, &a.AcquisitionDate, &a.AcquisitionType,
		&a.Supplier, &a.Location, &a.AssignedTo, &a.Notes,
		&a.IsActive, &a.CreatedAt, &a.UpdatedAt, &a.CreatedBy,
		&t.ID, &t.Name, &t.NameAr, &t.Unit,
		&c.ID, &c.Name, &c.NameAr, &c.Icon, &c.Color,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("asset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get asset: %w", err)
	}

	t.Category = &c
	a.AssetType = &t
	return &a, nil
}

// CreateAsset creates a new asset
func (s *AssetService) CreateAsset(ctx context.Context, req *models.CreateAssetRequest, userID uuid.UUID) (*models.Asset, error) {
	now := time.Now()

	// Generate asset code
	var count int
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM assets").Scan(&count)
	assetCode := fmt.Sprintf("AST-%d-%04d", now.Year(), count+1)

	quantity := req.Quantity
	if quantity == 0 {
		quantity = 1
	}
	totalValue := req.UnitPrice * float64(quantity)

	condition := req.Condition
	if condition == "" {
		condition = models.AssetConditionNew
	}

	currency := req.Currency
	if currency == "" {
		currency = "LYD"
	}

	acquisitionDate := now
	if req.AcquisitionDate != nil {
		acquisitionDate = *req.AcquisitionDate
	}

	asset := &models.Asset{
		ID:              uuid.New(),
		AssetTypeID:     req.AssetTypeID,
		ProjectID:       req.ProjectID,
		TransactionID:   req.TransactionID,
		AssetCode:       &assetCode,
		Quantity:        quantity,
		UnitPrice:       req.UnitPrice,
		TotalValue:      totalValue,
		Currency:        currency,
		Condition:       condition,
		Status:          models.AssetStatusInUse,
		AcquisitionDate: acquisitionDate,
		AcquisitionType: req.AcquisitionType,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       userID,
	}

	// Set optional fields
	if req.SerialNumber != "" {
		asset.SerialNumber = &req.SerialNumber
	}
	if req.Name != "" {
		asset.Name = &req.Name
	}
	if req.NameAr != "" {
		asset.NameAr = &req.NameAr
	}
	if req.Description != "" {
		asset.Description = &req.Description
	}
	if req.Supplier != "" {
		asset.Supplier = &req.Supplier
	}
	if req.Location != "" {
		asset.Location = &req.Location
	}
	if req.AssignedTo != nil {
		asset.AssignedTo = req.AssignedTo
	}
	if req.Notes != "" {
		asset.Notes = &req.Notes
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO assets (
			id, asset_type_id, project_id, transaction_id,
			asset_code, serial_number, name, name_ar, description,
			quantity, unit_price, total_value, currency,
			condition, status, acquisition_date, acquisition_type,
			supplier, location, assigned_to, notes,
			is_active, created_at, updated_at, created_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21,
			$22, $23, $24, $25
		)
	`, asset.ID, asset.AssetTypeID, asset.ProjectID, asset.TransactionID,
		asset.AssetCode, asset.SerialNumber, asset.Name, asset.NameAr, asset.Description,
		asset.Quantity, asset.UnitPrice, asset.TotalValue, asset.Currency,
		asset.Condition, asset.Status, asset.AcquisitionDate, asset.AcquisitionType,
		asset.Supplier, asset.Location, asset.AssignedTo, asset.Notes,
		asset.IsActive, asset.CreatedAt, asset.UpdatedAt, asset.CreatedBy)

	if err != nil {
		return nil, fmt.Errorf("failed to create asset: %w", err)
	}

	return asset, nil
}

// UpdateAsset updates an existing asset
func (s *AssetService) UpdateAsset(ctx context.Context, id uuid.UUID, req *models.UpdateAssetRequest) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE assets SET
			name = COALESCE($1, name),
			name_ar = COALESCE($2, name_ar),
			description = COALESCE($3, description),
			serial_number = COALESCE($4, serial_number),
			condition = COALESCE($5, condition),
			status = COALESCE($6, status),
			location = COALESCE($7, location),
			assigned_to = COALESCE($8, assigned_to),
			notes = COALESCE($9, notes),
			updated_at = $10
		WHERE id = $11
	`, req.Name, req.NameAr, req.Description, req.SerialNumber, req.Condition, req.Status, req.Location, req.AssignedTo, req.Notes, time.Now(), id)
	return err
}

// DeleteAsset soft deletes an asset
func (s *AssetService) DeleteAsset(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE assets SET is_active = false, updated_at = $1 WHERE id = $2
	`, time.Now(), id)
	return err
}

// GetInventorySummary returns a summary of assets by category
func (s *AssetService) GetInventorySummary(ctx context.Context) ([]map[string]interface{}, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.name, c.name_ar, c.icon, c.color,
		       COUNT(a.id) as asset_count,
		       COALESCE(SUM(a.total_value), 0) as total_value
		FROM asset_categories c
		LEFT JOIN asset_types t ON c.id = t.category_id
		LEFT JOIN assets a ON t.id = a.asset_type_id AND a.is_active = true
		WHERE c.is_active = true
		GROUP BY c.id, c.name, c.name_ar, c.icon, c.color, c.sort_order
		ORDER BY c.sort_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summary []map[string]interface{}
	for rows.Next() {
		var id uuid.UUID
		var name, nameAr, icon, color string
		var count int
		var total float64

		if err := rows.Scan(&id, &name, &nameAr, &icon, &color, &count, &total); err != nil {
			continue
		}

		summary = append(summary, map[string]interface{}{
			"id":          id,
			"name":        name,
			"name_ar":     nameAr,
			"icon":        icon,
			"color":       color,
			"asset_count": count,
			"total_value": total,
		})
	}

	return summary, nil
}
