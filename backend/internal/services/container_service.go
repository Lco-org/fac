package services

import (
	"context"
	"time"

	"cso-erp/internal/database"
	"cso-erp/internal/models"

	"github.com/google/uuid"
)

type ContainerService struct {
	db *database.DB
}

func NewContainerService(db *database.DB) *ContainerService {
	return &ContainerService{db: db}
}

// Common errors are defined in errors.go

// ListContainers returns all containers
func (s *ContainerService) ListContainers(ctx context.Context) ([]models.FinancialContainer, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar, 
		       account_number, iban, swift_code, location, location_ar, is_active, created_at
		FROM financial_containers
		ORDER BY type, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var containers []models.FinancialContainer
	for rows.Next() {
		var c models.FinancialContainer
		err := rows.Scan(
			&c.ID, &c.Name, &c.NameAr, &c.Type, &c.Currency, &c.Balance,
			&c.BankName, &c.BankNameAr, &c.AccountNumber, &c.IBAN, &c.SwiftCode,
			&c.Location, &c.LocationAr, &c.IsActive, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		containers = append(containers, c)
	}

	return containers, nil
}

// GetContainerByID retrieves a container by ID
func (s *ContainerService) GetContainerByID(ctx context.Context, id uuid.UUID) (*models.FinancialContainer, error) {
	var c models.FinancialContainer
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar, 
		       account_number, iban, swift_code, location, location_ar, is_active, created_at
		FROM financial_containers
		WHERE id = $1
	`, id).Scan(
		&c.ID, &c.Name, &c.NameAr, &c.Type, &c.Currency, &c.Balance,
		&c.BankName, &c.BankNameAr, &c.AccountNumber, &c.IBAN, &c.SwiftCode,
		&c.Location, &c.LocationAr, &c.IsActive, &c.CreatedAt,
	)

	if err != nil {
		return nil, ErrContainerNotFound
	}

	return &c, nil
}

// CreateContainer creates a new container
func (s *ContainerService) CreateContainer(ctx context.Context, req *models.FinancialContainer, createdBy uuid.UUID) (*models.FinancialContainer, error) {
	var c models.FinancialContainer
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO financial_containers (
			name, name_ar, type, currency, balance, bank_name, bank_name_ar,
			account_number, iban, swift_code, location, location_ar, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, name, name_ar, type, currency, balance, bank_name, bank_name_ar,
		          account_number, iban, swift_code, location, location_ar, is_active, created_at
	`,
		req.Name, req.NameAr, req.Type, req.Currency, req.Balance,
		req.BankName, req.BankNameAr, req.AccountNumber, req.IBAN, req.SwiftCode,
		req.Location, req.LocationAr, createdBy,
	).Scan(
		&c.ID, &c.Name, &c.NameAr, &c.Type, &c.Currency, &c.Balance,
		&c.BankName, &c.BankNameAr, &c.AccountNumber, &c.IBAN, &c.SwiftCode,
		&c.Location, &c.LocationAr, &c.IsActive, &c.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &c, nil
}

// UpdateContainer updates a container
func (s *ContainerService) UpdateContainer(ctx context.Context, id uuid.UUID, req *models.FinancialContainer) (*models.FinancialContainer, error) {
	var c models.FinancialContainer
	err := s.db.QueryRowContext(ctx, `
		UPDATE financial_containers SET
			name = $1, name_ar = $2, type = $3, currency = $4,
			bank_name = $5, bank_name_ar = $6, account_number = $7,
			iban = $8, swift_code = $9, location = $10, location_ar = $11,
			is_active = $12, updated_at = $13
		WHERE id = $14
		RETURNING id, name, name_ar, type, currency, balance, bank_name, bank_name_ar,
		          account_number, iban, swift_code, location, location_ar, is_active, created_at
	`,
		req.Name, req.NameAr, req.Type, req.Currency,
		req.BankName, req.BankNameAr, req.AccountNumber,
		req.IBAN, req.SwiftCode, req.Location, req.LocationAr,
		req.IsActive, time.Now(), id,
	).Scan(
		&c.ID, &c.Name, &c.NameAr, &c.Type, &c.Currency, &c.Balance,
		&c.BankName, &c.BankNameAr, &c.AccountNumber, &c.IBAN, &c.SwiftCode,
		&c.Location, &c.LocationAr, &c.IsActive, &c.CreatedAt,
	)

	if err != nil {
		return nil, ErrContainerNotFound
	}

	return &c, nil
}

// DeleteContainer deletes a container permanently along with all associated transactions
// WARNING: This is a CASCADE DELETE - all transactions will be permanently deleted
func (s *ContainerService) DeleteContainer(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM financial_containers WHERE id = $1
	`, id)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrContainerNotFound
	}

	return nil
}
