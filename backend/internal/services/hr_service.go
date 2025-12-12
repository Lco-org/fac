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

	deductions := []models.SalaryDeduction{}
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

	bonuses := []models.SalaryBonus{}
	for rows.Next() {
		var b models.SalaryBonus
		if err := rows.Scan(&b.ID, &b.EmployeeID, &b.MonthYear, &b.Amount, &b.Currency, &b.Reason, &b.ReasonAr, &b.CreatedAt, &b.UpdatedAt, &b.CreatedBy); err != nil {
			return nil, err
		}
		bonuses = append(bonuses, b)
	}

	return bonuses, nil
}

func (s *HRService) GetMonthlySalaries(ctx context.Context, monthYear time.Time, employeeID *uuid.UUID) ([]models.SalarySummary, error) {
	// 1. Get active employees (optionally filtered)
	query := `
		SELECT id, employee_number, full_name, base_salary, salary_currency 
		FROM employees 
		WHERE is_active = true
	`
	args := []interface{}{}

	if employeeID != nil {
		query += " AND id = $1"
		args = append(args, *employeeID)
	}

	query += " ORDER BY employee_number ASC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.SalarySummary

	// Create map for easy lookup if we wanted to optimization, but iterating is fine for now
	for rows.Next() {
		var s models.SalarySummary
		if err := rows.Scan(&s.EmployeeID, &s.EmployeeNumber, &s.EmployeeName, &s.BaseSalary, &s.Currency); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}

	// 2. Iterate each employee and calculate totals
	// Note: This matches the "Month" part of the date.
	startOfMonth := time.Date(monthYear.Year(), monthYear.Month(), 1, 0, 0, 0, 0, time.UTC)
	endOfMonth := startOfMonth.AddDate(0, 1, 0)

	for i := range summaries {
		empID := summaries[i].EmployeeID

		// Get Deductions
		var deductionTotal float64
		err := s.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(amount), 0) 
			FROM salary_deductions 
			WHERE employee_id = $1 AND month_year >= $2 AND month_year < $3
		`, empID, startOfMonth, endOfMonth).Scan(&deductionTotal)
		if err != nil {
			return nil, err
		}
		summaries[i].TotalDeductions = deductionTotal

		// Get Bonuses
		var bonusTotal float64
		err = s.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(amount), 0) 
			FROM salary_bonuses 
			WHERE employee_id = $1 AND month_year >= $2 AND month_year < $3
		`, empID, startOfMonth, endOfMonth).Scan(&bonusTotal)
		if err != nil {
			return nil, err
		}
		summaries[i].TotalBonuses = bonusTotal

		// Verify Currency match? Assuming all in same currency or base currency for now.
		// If base salary is 0, Net might be negative? Application logic:
		summaries[i].NetSalary = summaries[i].BaseSalary + summaries[i].TotalBonuses - summaries[i].TotalDeductions
	}

	return summaries, nil
}
