package services

import (
	"context"
	"time"

	"cso-erp/internal/database"
	"cso-erp/internal/models"

	"github.com/google/uuid"
)

type HRService struct {
	db *database.DB
}

func NewHRService(db *database.DB) *HRService {
	return &HRService{db: db}
}

// ============================================================================
// DEDUCTIONS
// ============================================================================

func (s *HRService) CreateDeduction(ctx context.Context, req *models.CreateSalaryDeductionRequest, createdBy uuid.UUID) (*models.SalaryDeduction, error) {
	deduction := &models.SalaryDeduction{
		ID:         uuid.New(),
		EmployeeID: req.EmployeeID,
		MonthYear:  req.MonthYear,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Reason:     req.Reason,
		ReasonAr:   req.ReasonAr,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		CreatedBy:  &createdBy,
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO salary_deductions (
			id, employee_id, month_year, amount, currency, reason, reason_ar, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, deduction.ID, deduction.EmployeeID, deduction.MonthYear, deduction.Amount, deduction.Currency, deduction.Reason, deduction.ReasonAr, deduction.CreatedAt, deduction.UpdatedAt, deduction.CreatedBy)

	if err != nil {
		return nil, err
	}

	return deduction, nil
}

func (s *HRService) ListDeductions(ctx context.Context, employeeID *uuid.UUID) ([]models.SalaryDeduction, error) {
	query := `
		SELECT id, employee_id, month_year, amount, currency, reason, reason_ar, created_at, updated_at, created_by
		FROM salary_deductions
	`
	args := []interface{}{}

	if employeeID != nil {
		query += " WHERE employee_id = $1"
		args = append(args, *employeeID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var deductions []models.SalaryDeduction
	for rows.Next() {
		var d models.SalaryDeduction
		if err := rows.Scan(&d.ID, &d.EmployeeID, &d.MonthYear, &d.Amount, &d.Currency, &d.Reason, &d.ReasonAr, &d.CreatedAt, &d.UpdatedAt, &d.CreatedBy); err != nil {
			return nil, err
		}
		deductions = append(deductions, d)
	}

	return deductions, nil
}

// ============================================================================
// BONUSES
// ============================================================================

func (s *HRService) CreateBonus(ctx context.Context, req *models.CreateSalaryBonusRequest, createdBy uuid.UUID) (*models.SalaryBonus, error) {
	bonus := &models.SalaryBonus{
		ID:         uuid.New(),
		EmployeeID: req.EmployeeID,
		MonthYear:  req.MonthYear,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Reason:     req.Reason,
		ReasonAr:   req.ReasonAr,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		CreatedBy:  &createdBy,
	}

	// Assuming salary_bonuses table exists
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO salary_bonuses (
			id, employee_id, month_year, amount, currency, reason, reason_ar, created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, bonus.ID, bonus.EmployeeID, bonus.MonthYear, bonus.Amount, bonus.Currency, bonus.Reason, bonus.ReasonAr, bonus.CreatedAt, bonus.UpdatedAt, bonus.CreatedBy)

	if err != nil {
		return nil, err
	}

	return bonus, nil
}

func (s *HRService) ListBonuses(ctx context.Context, employeeID *uuid.UUID) ([]models.SalaryBonus, error) {
	query := `
		SELECT id, employee_id, month_year, amount, currency, reason, reason_ar, created_at, updated_at, created_by
		FROM salary_bonuses
	`
	args := []interface{}{}

	if employeeID != nil {
		query += " WHERE employee_id = $1"
		args = append(args, *employeeID)
	}

	query += " ORDER BY created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bonuses []models.SalaryBonus
	for rows.Next() {
		var b models.SalaryBonus
		if err := rows.Scan(&b.ID, &b.EmployeeID, &b.MonthYear, &b.Amount, &b.Currency, &b.Reason, &b.ReasonAr, &b.CreatedAt, &b.UpdatedAt, &b.CreatedBy); err != nil {
			return nil, err
		}
		bonuses = append(bonuses, b)
	}

	return bonuses, nil
}
