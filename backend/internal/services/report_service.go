package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"time"
)

type ReportService struct {
	db *database.DB
}

func NewReportService(db *database.DB) *ReportService {
	return &ReportService{db: db}
}

// GetContainerStatement retrieves detailed financial statement for a specific container
func (s *ReportService) GetContainerStatement(ctx context.Context, containerID string, startDate, endDate *time.Time) (*models.ContainerStatement, error) {
	// Get container details
	var container models.FinancialContainer
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar, 
		       account_number, iban, location, location_ar, is_active, created_at
		FROM financial_containers
		WHERE id = $1
	`, containerID).Scan(
		&container.ID, &container.Name, &container.NameAr, &container.Type, &container.Currency,
		&container.Balance, &container.BankName, &container.BankNameAr, &container.AccountNumber,
		&container.IBAN, &container.Location, &container.LocationAr, &container.IsActive, &container.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Build query with date filters
	query := `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency, 
		       t.description, t.description_ar, t.payee_payer,
		       p.code, p.name, t.balance_before, t.balance_after,
		       t.transaction_date, t.created_at
		FROM transactions t
		JOIN projects p ON t.project_id = p.id
		WHERE t.container_id = $1 AND t.status = 'completed'
	`
	args := []interface{}{containerID}
	argCount := 1

	if startDate != nil {
		argCount++
		query += ` AND t.transaction_date >= $` + string(rune(argCount+'0'))
		args = append(args, startDate)
	}
	if endDate != nil {
		argCount++
		query += ` AND t.transaction_date <= $` + string(rune(argCount+'0'))
		args = append(args, endDate)
	}

	query += ` ORDER BY t.transaction_date ASC, t.created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionDetail
	var totalIncome, totalExpenses float64
	var openingBalance, closingBalance float64
	firstTransaction := true

	for rows.Next() {
		var td models.TransactionDetail
		err := rows.Scan(
			&td.ID, &td.TransactionNumber, &td.Type, &td.Amount, &td.Currency,
			&td.Description, &td.DescriptionAr, &td.PayeePayer,
			&td.ProjectCode, &td.ProjectName, &td.BalanceBefore, &td.BalanceAfter,
			&td.TransactionDate, &td.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if firstTransaction {
			openingBalance = td.BalanceBefore
			firstTransaction = false
		}
		closingBalance = td.BalanceAfter

		if td.Type == "income_grant" || td.Type == "income_donation" {
			totalIncome += td.Amount
		} else {
			totalExpenses += td.Amount
		}

		transactions = append(transactions, td)
	}

	if len(transactions) == 0 {
		openingBalance = container.Balance
		closingBalance = container.Balance
	}

	statement := &models.ContainerStatement{
		Container:    container,
		Transactions: transactions,
		Summary: models.ContainerStatementSummary{
			OpeningBalance:   openingBalance,
			TotalIncome:      totalIncome,
			TotalExpenses:    totalExpenses,
			ClosingBalance:   closingBalance,
			TransactionCount: len(transactions),
		},
		GeneratedAt: time.Now(),
	}

	return statement, nil
}

// GetProjectStatement retrieves detailed financial statement for a specific project
func (s *ReportService) GetProjectStatement(ctx context.Context, projectID string, startDate, endDate *time.Time) (*models.ProjectStatement, error) {
	// Get project details
	var project models.Project
	err := s.db.QueryRowContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active, created_at
		FROM projects
		WHERE id = $1
	`, projectID).Scan(
		&project.ID, &project.Code, &project.Name, &project.NameAr, &project.Description,
		&project.DonorName, &project.DonorNameAr, &project.StartDate, &project.EndDate,
		&project.Budget, &project.Currency, &project.IsActive, &project.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Build query with date filters
	query := `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar, t.payee_payer,
		       fc.name, fc.type, t.transaction_date, t.created_at
		FROM transactions t
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.project_id = $1 AND t.status = 'completed'
	`
	args := []interface{}{projectID}
	argCount := 1

	if startDate != nil {
		argCount++
		query += ` AND t.transaction_date >= $` + string(rune(argCount+'0'))
		args = append(args, startDate)
	}
	if endDate != nil {
		argCount++
		query += ` AND t.transaction_date <= $` + string(rune(argCount+'0'))
		args = append(args, endDate)
	}

	query += ` ORDER BY t.transaction_date ASC, t.created_at ASC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.ProjectTransactionDetail
	var totalIncome, totalExpenses float64

	for rows.Next() {
		var ptd models.ProjectTransactionDetail
		err := rows.Scan(
			&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
			&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
			&ptd.ContainerName, &ptd.ContainerType, &ptd.TransactionDate, &ptd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if ptd.Type == "income_grant" || ptd.Type == "income_donation" {
			totalIncome += ptd.Amount
		} else {
			totalExpenses += ptd.Amount
		}

		transactions = append(transactions, ptd)
	}

	netPosition := totalIncome - totalExpenses
	var budgetUtilization float64
	if project.Budget != nil && *project.Budget > 0 {
		budgetUtilization = (totalExpenses / *project.Budget) * 100
	}

	budget := 0.0
	if project.Budget != nil {
		budget = *project.Budget
	}

	statement := &models.ProjectStatement{
		Project:      project,
		Transactions: transactions,
		Summary: models.ProjectStatementSummary{
			TotalIncome:       totalIncome,
			TotalExpenses:     totalExpenses,
			NetPosition:       netPosition,
			TransactionCount:  len(transactions),
			Budget:            budget,
			BudgetUtilization: budgetUtilization,
		},
		GeneratedAt: time.Now(),
	}

	return statement, nil
}
