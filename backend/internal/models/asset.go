package models

import (
	"time"

	"github.com/google/uuid"
)

// AssetCategory represents a category of assets (e.g., Electronics, Furniture)
type AssetCategory struct {
	ID        uuid.UUID `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	NameAr    string    `json:"name_ar" db:"name_ar"`
	Icon      string    `json:"icon" db:"icon"`
	Color     string    `json:"color" db:"color"`
	SortOrder int       `json:"sort_order" db:"sort_order"`
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AssetType represents a type within a category (e.g., Laptop, Printer)
type AssetType struct {
	ID                uuid.UUID `json:"id" db:"id"`
	CategoryID        uuid.UUID `json:"category_id" db:"category_id"`
	Name              string    `json:"name" db:"name"`
	NameAr            string    `json:"name_ar" db:"name_ar"`
	Unit              string    `json:"unit" db:"unit"`
	DepreciationYears *int      `json:"depreciation_years,omitempty" db:"depreciation_years"`
	IsActive          bool      `json:"is_active" db:"is_active"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`

	// Joined data
	Category *AssetCategory `json:"category,omitempty"`
}

// Asset represents a physical asset owned by the organization
type Asset struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	AssetTypeID   uuid.UUID  `json:"asset_type_id" db:"asset_type_id"`
	ProjectID     *uuid.UUID `json:"project_id,omitempty" db:"project_id"`
	TransactionID *uuid.UUID `json:"transaction_id,omitempty" db:"transaction_id"`

	// Identification
	AssetCode    *string `json:"asset_code,omitempty" db:"asset_code"`
	SerialNumber *string `json:"serial_number,omitempty" db:"serial_number"`
	Barcode      *string `json:"barcode,omitempty" db:"barcode"`

	// Description
	Name          *string `json:"name,omitempty" db:"name"`
	NameAr        *string `json:"name_ar,omitempty" db:"name_ar"`
	Description   *string `json:"description,omitempty" db:"description"`
	DescriptionAr *string `json:"description_ar,omitempty" db:"description_ar"`

	// Value
	Quantity   int     `json:"quantity" db:"quantity"`
	UnitPrice  float64 `json:"unit_price" db:"unit_price"`
	TotalValue float64 `json:"total_value" db:"total_value"`
	Currency   string  `json:"currency" db:"currency"`

	// Status
	Condition string `json:"condition" db:"condition"`
	Status    string `json:"status" db:"status"`

	// Tracking
	AcquisitionDate time.Time  `json:"acquisition_date" db:"acquisition_date"`
	AcquisitionType string     `json:"acquisition_type" db:"acquisition_type"`
	Supplier        *string    `json:"supplier,omitempty" db:"supplier"`
	WarrantyExpires *time.Time `json:"warranty_expires,omitempty" db:"warranty_expires"`

	// Location
	Location   *string    `json:"location,omitempty" db:"location"`
	AssignedTo *uuid.UUID `json:"assigned_to,omitempty" db:"assigned_to"`

	// Notes
	Notes *string `json:"notes,omitempty" db:"notes"`

	// System
	IsActive  bool      `json:"is_active" db:"is_active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy uuid.UUID `json:"created_by" db:"created_by"`

	// Joined data
	AssetType      *AssetType `json:"asset_type,omitempty"`
	AssignedToName *string    `json:"assigned_to_name,omitempty"`
}

// Asset condition constants
const (
	AssetConditionNew     = "new"
	AssetConditionGood    = "good"
	AssetConditionFair    = "fair"
	AssetConditionPoor    = "poor"
	AssetConditionDamaged = "damaged"
)

// Asset status constants
const (
	AssetStatusInUse       = "in_use"
	AssetStatusInStorage   = "in_storage"
	AssetStatusDisposed    = "disposed"
	AssetStatusTransferred = "transferred"
)

// Acquisition type constants
const (
	AcquisitionTypeDonation = "donation"
	AcquisitionTypePurchase = "purchase"
	AcquisitionTypeGrant    = "grant"
	AcquisitionTypeTransfer = "transfer"
)

// CreateAssetCategoryRequest represents the request to create an asset category
type CreateAssetCategoryRequest struct {
	Name      string `json:"name" validate:"required"`
	NameAr    string `json:"name_ar" validate:"required"`
	Icon      string `json:"icon"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

// CreateAssetTypeRequest represents the request to create an asset type
type CreateAssetTypeRequest struct {
	CategoryID        uuid.UUID `json:"category_id" validate:"required"`
	Name              string    `json:"name" validate:"required"`
	NameAr            string    `json:"name_ar" validate:"required"`
	Unit              string    `json:"unit"`
	DepreciationYears *int      `json:"depreciation_years"`
}

// CreateAssetRequest represents the request to create an asset
type CreateAssetRequest struct {
	AssetTypeID   uuid.UUID  `json:"asset_type_id" validate:"required"`
	ProjectID     *uuid.UUID `json:"project_id"`
	TransactionID *uuid.UUID `json:"transaction_id"`

	SerialNumber  string `json:"serial_number"`
	Name          string `json:"name"`
	NameAr        string `json:"name_ar"`
	Description   string `json:"description"`
	DescriptionAr string `json:"description_ar"`

	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
	Currency  string  `json:"currency"`

	Condition       string     `json:"condition"`
	AcquisitionDate *time.Time `json:"acquisition_date"`
	AcquisitionType string     `json:"acquisition_type" validate:"required"`
	Supplier        string     `json:"supplier"`
	Location        string     `json:"location"`
	AssignedTo      *uuid.UUID `json:"assigned_to"`
	Notes           string     `json:"notes"`
}

// UpdateAssetRequest represents the request to update an asset
type UpdateAssetRequest struct {
	Name         *string    `json:"name"`
	NameAr       *string    `json:"name_ar"`
	Description  *string    `json:"description"`
	SerialNumber *string    `json:"serial_number"`
	Condition    *string    `json:"condition"`
	Status       *string    `json:"status"`
	Location     *string    `json:"location"`
	AssignedTo   *uuid.UUID `json:"assigned_to"`
	Notes        *string    `json:"notes"`
}

// AssetInput represents asset data when creating a transaction with assets
type AssetInput struct {
	AssetTypeID   uuid.UUID `json:"asset_type_id" validate:"required"`
	Quantity      int       `json:"quantity"`
	UnitPrice     float64   `json:"unit_price"`
	SerialNumber  string    `json:"serial_number"`
	Name          string    `json:"name"`
	NameAr        string    `json:"name_ar"`
	Description   string    `json:"description"`
	DescriptionAr string    `json:"description_ar"`
	Condition     string    `json:"condition"`
	Location      string    `json:"location"`
}

// AssetListFilter represents filters for listing assets
type AssetListFilter struct {
	CategoryID      *uuid.UUID `query:"category_id"`
	TypeID          *uuid.UUID `query:"type_id"`
	ProjectID       *uuid.UUID `query:"project_id"`
	Status          string     `query:"status"`
	Condition       string     `query:"condition"`
	AcquisitionType string     `query:"acquisition_type"`
	AssignedTo      *uuid.UUID `query:"assigned_to"`
	Search          string     `query:"search"`
	Page            int        `query:"page"`
	Limit           int        `query:"limit"`
}
