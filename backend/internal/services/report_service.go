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

// GetDashboardSummary retrieves the dashboard summary with detailed transactions
func (s *ReportService) GetDashboardSummary(ctx context.Context) (*models.DashboardSummary, error) {
	// 1. Get all active bank containers
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar, 
		       account_number, iban, location, location_ar, is_active, created_at
		FROM financial_containers
		WHERE type = 'bank' AND is_active = true
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bankSummaries []models.BankDashboardSummary

	for rows.Next() {
		var c models.FinancialContainer
		err := rows.Scan(
			&c.ID, &c.Name, &c.NameAr, &c.Type, &c.Currency,
			&c.Balance, &c.BankName, &c.BankNameAr, &c.AccountNumber,
			&c.IBAN, &c.Location, &c.LocationAr, &c.IsActive, &c.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// 2. For each bank, get last 3 transactions
		txRows, err := s.db.QueryContext(ctx, `
			SELECT t.id, t.transaction_number, t.type, t.amount, t.currency, 
			       t.description, t.description_ar, t.payee_payer,
			       p.code, p.name, t.balance_before, t.balance_after,
			       t.transaction_date, t.created_at
			FROM transactions t
			JOIN projects p ON t.project_id = p.id
			WHERE t.container_id = $1 AND t.status = 'completed'
			ORDER BY t.transaction_date DESC, t.created_at DESC
			LIMIT 3
		`, c.ID)
		if err != nil {
			return nil, err
		}
		defer txRows.Close()

		var transactions []models.TransactionDetail
		for txRows.Next() {
			var td models.TransactionDetail
			err := txRows.Scan(
				&td.ID, &td.TransactionNumber, &td.Type, &td.Amount, &td.Currency,
				&td.Description, &td.DescriptionAr, &td.PayeePayer,
				&td.ProjectCode, &td.ProjectName, &td.BalanceBefore, &td.BalanceAfter,
				&td.TransactionDate, &td.CreatedAt,
			)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, td)
		}

		bankSummaries = append(bankSummaries, models.BankDashboardSummary{
			Container:        c,
			LastTransactions: transactions,
		})
	}

	// 3. Get active projects with last 3 transactions
	projectRows, err := s.db.QueryContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active, created_at
		FROM projects
		WHERE is_active = true
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer projectRows.Close()

	var projectSummaries []models.ProjectDashboardSummary

	for projectRows.Next() {
		var p models.Project
		err := projectRows.Scan(
			&p.ID, &p.Code, &p.Name, &p.NameAr, &p.Description,
			&p.DonorName, &p.DonorNameAr, &p.StartDate, &p.EndDate,
			&p.Budget, &p.Currency, &p.IsActive, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// For each project, get last 3 transactions
		ptxRows, err := s.db.QueryContext(ctx, `
			SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
			       t.description, t.description_ar, t.payee_payer,
			       fc.name, fc.type, t.transaction_date, t.created_at
			FROM transactions t
			JOIN financial_containers fc ON t.container_id = fc.id
			WHERE t.project_id = $1 AND t.status = 'completed'
			ORDER BY t.transaction_date DESC, t.created_at DESC
			LIMIT 3
		`, p.ID)
		if err != nil {
			return nil, err
		}
		defer ptxRows.Close()

		var transactions []models.ProjectTransactionDetail
		for ptxRows.Next() {
			var ptd models.ProjectTransactionDetail
			err := ptxRows.Scan(
				&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
				&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
				&ptd.ContainerName, &ptd.ContainerType, &ptd.TransactionDate, &ptd.CreatedAt,
			)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, ptd)
		}

		projectSummaries = append(projectSummaries, models.ProjectDashboardSummary{
			Project:          p,
			LastTransactions: transactions,
		})
	}

	// 4. Get recent income (global)
	incomeRows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar, t.payee_payer,
		       fc.name, fc.type, t.transaction_date, t.created_at
		FROM transactions t
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.type IN ('income_grant', 'income_donation') AND t.status = 'completed'
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT 3
	`)
	if err != nil {
		return nil, err
	}
	defer incomeRows.Close()

	var recentIncome []models.ProjectTransactionDetail
	for incomeRows.Next() {
		var ptd models.ProjectTransactionDetail
		err := incomeRows.Scan(
			&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
			&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
			&ptd.ContainerName, &ptd.ContainerType, &ptd.TransactionDate, &ptd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		recentIncome = append(recentIncome, ptd)
	}

	// 4. Get recent expenses (global)
	expenseRows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar, t.payee_payer,
		       fc.name, fc.type, t.transaction_date, t.created_at
		FROM transactions t
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.type NOT IN ('income_grant', 'income_donation', 'transfer') AND t.status = 'completed'
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT 3
	`)
	if err != nil {
		return nil, err
	}
	defer expenseRows.Close()

	var recentExpenses []models.ProjectTransactionDetail
	for expenseRows.Next() {
		var ptd models.ProjectTransactionDetail
		err := expenseRows.Scan(
			&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
			&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
			&ptd.ContainerName, &ptd.ContainerType, &ptd.TransactionDate, &ptd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		recentExpenses = append(recentExpenses, ptd)
	}

	return &models.DashboardSummary{
		BankAccounts:   bankSummaries,
		Projects:       projectSummaries,
		RecentIncome:   recentIncome,
		RecentExpenses: recentExpenses,
		GeneratedAt:    time.Now(),
	}, nil
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
