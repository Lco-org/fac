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

// GetDailyPosition retrieves the daily financial position report
func (s *ReportService) GetDailyPosition(ctx context.Context) (*models.DailyPositionReport, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, balance_status,
			   today_income, today_expenses, bank_name, account_number, location
		FROM v_daily_position
		ORDER BY type, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report models.DailyPositionReport
	var cashContainers []models.ContainerPosition
	var bankContainers []models.ContainerPosition
	var totalCash, totalBank, totalDeficit float64

	for rows.Next() {
		var cp models.ContainerPosition
		err := rows.Scan(
			&cp.ID, &cp.Name, &cp.NameAr, &cp.Type, &cp.Currency,
			&cp.Balance, &cp.BalanceStatus, &cp.TodayIncome, &cp.TodayExpenses,
			&cp.BankName, &cp.AccountNumber, &cp.Location,
		)
		if err != nil {
			return nil, err
		}

		if cp.Type == models.ContainerCash {
			cashContainers = append(cashContainers, cp)
			totalCash += cp.Balance
		} else {
			bankContainers = append(bankContainers, cp)
			totalBank += cp.Balance
		}

		if cp.Balance < 0 {
			totalDeficit += cp.Balance
		}
	}

	report.CashContainers = cashContainers
	report.BankContainers = bankContainers
	report.TotalCash = totalCash
	report.TotalBank = totalBank
	report.TotalDeficit = totalDeficit
	report.GeneratedAt = time.Now()

	return &report, nil
}

// GetProjectSummary retrieves project financial summary
func (s *ReportService) GetProjectSummary(ctx context.Context) ([]models.ProjectSummary, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, code, name, name_ar, currency, donor_name,
			   total_income, total_expenses, net_position, transaction_count
		FROM v_project_summary
		ORDER BY code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []models.ProjectSummary
	for rows.Next() {
		var p models.ProjectSummary
		err := rows.Scan(
			&p.ID, &p.Code, &p.Name, &p.NameAr, &p.Currency, &p.DonorName,
			&p.TotalIncome, &p.TotalExpenses, &p.NetPosition, &p.TransactionCount,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

// GetDeficitReport retrieves containers with deficit
func (s *ReportService) GetDeficitReport(ctx context.Context) (*models.DeficitReport, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, deficit_amount,
			   deficit_transactions_count, first_deficit_date, last_deficit_date
		FROM v_deficit_report
		ORDER BY deficit_amount ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var report models.DeficitReport
	var containers []models.DeficitContainer
	var totalDeficit float64

	for rows.Next() {
		var dc models.DeficitContainer
		err := rows.Scan(
			&dc.ID, &dc.Name, &dc.NameAr, &dc.Type, &dc.Currency,
			&dc.DeficitAmount, &dc.DeficitTransactionsCount,
			&dc.FirstDeficitDate, &dc.LastDeficitDate,
		)
		if err != nil {
			return nil, err
		}
		containers = append(containers, dc)
		totalDeficit += dc.DeficitAmount
	}

	report.Containers = containers
	report.TotalDeficit = totalDeficit
	report.ContainersCount = len(containers)
	report.GeneratedAt = time.Now()

	return &report, nil
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
