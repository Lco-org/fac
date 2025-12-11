package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"log"
	"time"
)

type ReportService struct {
	db *database.DB
}

func NewReportService(db *database.DB) *ReportService {
	return &ReportService{db: db}
}

func (s *ReportService) GetDashboardSummary(ctx context.Context, userRole string) (*models.DashboardSummary, error) {
	dashService := NewDashboardService(s.db)
	visibleWidgets, err := dashService.GetVisibleWidgets(ctx, userRole)
	if err != nil {
		return nil, err
	}

	widgetMap := make(map[string]bool)
	for _, w := range visibleWidgets {
		widgetMap[w.WidgetKey] = true
	}

	bankSummaries := make([]models.BankDashboardSummary, 0)
	projectSummaries := make([]models.ProjectDashboardSummary, 0)
	recentIncome := make([]models.ProjectTransactionDetail, 0)
	recentExpenses := make([]models.ProjectTransactionDetail, 0)

	if widgetMap["banks"] {
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

			txRows, err := s.db.QueryContext(ctx, `
			SELECT t.id, t.transaction_number, t.type, t.amount, t.currency, 
			       t.description, t.description_ar, t.payee_payer,
			       COALESCE(p.code, ''), COALESCE(p.name, ''), t.balance_before, t.balance_after,
			       t.transaction_date, t.created_at
			FROM transactions t
			LEFT JOIN projects p ON t.project_id = p.id
			WHERE t.container_id = $1 AND t.status = 'completed'
			ORDER BY t.transaction_date DESC, t.created_at DESC
			LIMIT 3
		`, c.ID)
			if err != nil {
				return nil, err
			}
			defer txRows.Close()

			transactions := make([]models.TransactionDetail, 0)
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
				td.Currency = c.Currency
				transactions = append(transactions, td)
			}

			bankSummaries = append(bankSummaries, models.BankDashboardSummary{
				Container:        c,
				LastTransactions: transactions,
			})
		}
	}

	cashSummaries := make([]models.BankDashboardSummary, 0)
	if widgetMap["cash_accounts"] {
		rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar, 
		       account_number, iban, location, location_ar, is_active, created_at
		FROM financial_containers
		WHERE type = 'cash' AND is_active = true
		ORDER BY name
	`)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

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

			txRows, err := s.db.QueryContext(ctx, `
			SELECT t.id, t.transaction_number, t.type, t.amount, t.currency, 
			       t.description, t.description_ar, t.payee_payer,
			       COALESCE(p.code, ''), COALESCE(p.name, ''), t.balance_before, t.balance_after,
			       t.transaction_date, t.created_at
			FROM transactions t
			LEFT JOIN projects p ON t.project_id = p.id
			WHERE t.container_id = $1 AND t.status = 'completed'
			ORDER BY t.transaction_date DESC, t.created_at DESC
			LIMIT 3
		`, c.ID)
			if err != nil {
				return nil, err
			}
			defer txRows.Close()

			transactions := make([]models.TransactionDetail, 0)
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
				td.Currency = c.Currency
				transactions = append(transactions, td)
			}

			cashSummaries = append(cashSummaries, models.BankDashboardSummary{
				Container:        c,
				LastTransactions: transactions,
			})
		}
	}

	if widgetMap["projects"] {
		projectRows, err := s.db.QueryContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active,
		       is_locked, locked_at, lock_reason, locked_by, created_at
		FROM projects
		WHERE is_active = true
		ORDER BY name
	`)
		if err != nil {
			return nil, err
		}
		defer projectRows.Close()

		for projectRows.Next() {
			var p models.Project
			err := projectRows.Scan(
				&p.ID, &p.Code, &p.Name, &p.NameAr, &p.Description,
				&p.DonorName, &p.DonorNameAr, &p.StartDate, &p.EndDate,
				&p.Budget, &p.Currency, &p.IsActive,
				&p.IsLocked, &p.LockedAt, &p.LockReason, &p.LockedBy, &p.CreatedAt,
			)
			if err != nil {
				return nil, err
			}

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

			transactions := make([]models.ProjectTransactionDetail, 0)
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
				ptd.Currency = p.Currency
				transactions = append(transactions, ptd)
			}

			projectSummaries = append(projectSummaries, models.ProjectDashboardSummary{
				Project:          p,
				LastTransactions: transactions,
			})
		}
	}

	if widgetMap["recent_income"] {
		incomeRows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar, t.payee_payer,
		       COALESCE(p.code, ''), COALESCE(p.name, ''),
		       fc.name, fc.type, fc.currency, t.transaction_date, t.created_at
		FROM transactions t
		LEFT JOIN projects p ON t.project_id = p.id
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.type IN ('income_grant', 'income_donation') AND t.status = 'completed'
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT 3
	`)
		if err != nil {
			return nil, err
		}
		defer incomeRows.Close()

		for incomeRows.Next() {
			var ptd models.ProjectTransactionDetail
			var containerCurrency models.CurrencyCode
			err := incomeRows.Scan(
				&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
				&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
				&ptd.ProjectCode, &ptd.ProjectName,
				&ptd.ContainerName, &ptd.ContainerType, &containerCurrency, &ptd.TransactionDate, &ptd.CreatedAt,
			)
			if err != nil {
				return nil, err
			}
			ptd.Currency = containerCurrency
			recentIncome = append(recentIncome, ptd)
		}
	}

	if widgetMap["recent_expenses"] {
		expenseRows, err := s.db.QueryContext(ctx, `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar, t.payee_payer,
		       COALESCE(p.code, ''), COALESCE(p.name, ''),
		       fc.name, fc.type, fc.currency, t.transaction_date, t.created_at
		FROM transactions t
		LEFT JOIN projects p ON t.project_id = p.id
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.type::text LIKE 'disbursement_%' AND t.status = 'completed'
		ORDER BY t.transaction_date DESC, t.created_at DESC
		LIMIT 3
	`)
		if err != nil {
			return nil, err
		}
		defer expenseRows.Close()

		for expenseRows.Next() {
			var ptd models.ProjectTransactionDetail
			var containerCurrency models.CurrencyCode
			err := expenseRows.Scan(
				&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
				&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
				&ptd.ProjectCode, &ptd.ProjectName,
				&ptd.ContainerName, &ptd.ContainerType, &containerCurrency, &ptd.TransactionDate, &ptd.CreatedAt,
			)
			if err != nil {
				return nil, err
			}
			ptd.Currency = containerCurrency
			recentExpenses = append(recentExpenses, ptd)
		}
	}

	var summary models.FinancialSummary
	if widgetMap["financial_summary"] {
		var err error
		summary, err = s.calculateFinancialSummary(ctx)
		if err != nil {
			log.Printf("ERROR: Failed to calculate financial summary: %v", err)
			summary = models.FinancialSummary{
				TotalBalance:  make([]models.CurrencyAmount, 0),
				TotalIncome:   make([]models.CurrencyAmount, 0),
				TotalExpenses: make([]models.CurrencyAmount, 0),
				TotalCustody:  make([]models.CurrencyAmount, 0),
			}
		}
	} else {
		summary = models.FinancialSummary{
			TotalBalance:  make([]models.CurrencyAmount, 0),
			TotalIncome:   make([]models.CurrencyAmount, 0),
			TotalExpenses: make([]models.CurrencyAmount, 0),
			TotalCustody:  make([]models.CurrencyAmount, 0),
		}
	}

	return &models.DashboardSummary{
		Summary:        summary,
		BankAccounts:   bankSummaries,
		CashAccounts:   cashSummaries,
		Projects:       projectSummaries,
		RecentIncome:   recentIncome,
		RecentExpenses: recentExpenses,
		GeneratedAt:    time.Now(),
	}, nil
}

func (s *ReportService) calculateFinancialSummary(ctx context.Context) (models.FinancialSummary, error) {
	summary := models.FinancialSummary{
		TotalBalance:  make([]models.CurrencyAmount, 0),
		TotalIncome:   make([]models.CurrencyAmount, 0),
		TotalExpenses: make([]models.CurrencyAmount, 0),
		TotalCustody:  make([]models.CurrencyAmount, 0),
	}

	balanceRows, err := s.db.QueryContext(ctx, `
		SELECT currency, SUM(balance) as total
		FROM financial_containers
		WHERE is_active = true
		GROUP BY currency
		ORDER BY currency
	`)
	if err != nil {
		return summary, err
	}
	defer balanceRows.Close()

	for balanceRows.Next() {
		var ca models.CurrencyAmount
		if err := balanceRows.Scan(&ca.Currency, &ca.Amount); err != nil {
			continue
		}
		summary.TotalBalance = append(summary.TotalBalance, ca)
	}

	incomeRows, err := s.db.QueryContext(ctx, `
		SELECT fc.currency, SUM(t.amount) as total
		FROM transactions t
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.type IN ('income_grant', 'income_donation') 
		  AND t.status = 'completed'
		GROUP BY fc.currency
		ORDER BY fc.currency
	`)
	if err != nil {
		return summary, err
	}
	defer incomeRows.Close()

	for incomeRows.Next() {
		var ca models.CurrencyAmount
		if err := incomeRows.Scan(&ca.Currency, &ca.Amount); err != nil {
			continue
		}
		summary.TotalIncome = append(summary.TotalIncome, ca)
	}

	expenseRows, err := s.db.QueryContext(ctx, `
		SELECT fc.currency, SUM(t.amount) as total
		FROM transactions t
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.type::text LIKE 'disbursement_%' 
		  AND t.status = 'completed'
		GROUP BY fc.currency
		ORDER BY fc.currency
	`)
	if err != nil {
		return summary, err
	}
	defer expenseRows.Close()

	for expenseRows.Next() {
		var ca models.CurrencyAmount
		if err := expenseRows.Scan(&ca.Currency, &ca.Amount); err != nil {
			continue
		}
		summary.TotalExpenses = append(summary.TotalExpenses, ca)
	}

	custodyRows, err := s.db.QueryContext(ctx, `
		SELECT fc.currency, SUM(t.amount) as total
		FROM transactions t
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.type = 'disbursement_custody' 
		  AND t.status = 'completed'
		GROUP BY fc.currency
		ORDER BY fc.currency
	`)
	if err != nil {
		return summary, err
	}
	defer custodyRows.Close()

	for custodyRows.Next() {
		var ca models.CurrencyAmount
		if err := custodyRows.Scan(&ca.Currency, &ca.Amount); err != nil {
			continue
		}
		summary.TotalCustody = append(summary.TotalCustody, ca)
	}

	return summary, nil
}

func (s *ReportService) GetContainerStatement(ctx context.Context, containerID string, startDate, endDate *time.Time) (*models.ContainerStatement, error) {
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

	query := `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency, 
		       t.description, t.description_ar, t.payee_payer,
		       COALESCE(p.code, ''), COALESCE(p.name, ''), t.balance_before, t.balance_after,
		       t.transaction_date, t.created_at
		FROM transactions t
		LEFT JOIN projects p ON t.project_id = p.id
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

	log.Printf("[DEBUG] GetContainerStatement: containerID=%s, query=%s, args=%v", containerID, query, args)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		log.Printf("[ERROR] GetContainerStatement query failed: %v", err)
		return nil, err
	}
	defer rows.Close()

	transactions := make([]models.TransactionDetail, 0)
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

func (s *ReportService) GetProjectStatement(ctx context.Context, projectID string, startDate, endDate *time.Time) (*models.ProjectStatement, error) {
	var project models.Project
	err := s.db.QueryRowContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active,
		       is_locked, locked_at, lock_reason, locked_by, created_at
		FROM projects
		WHERE id = $1
	`, projectID).Scan(
		&project.ID, &project.Code, &project.Name, &project.NameAr, &project.Description,
		&project.DonorName, &project.DonorNameAr, &project.StartDate, &project.EndDate,
		&project.Budget, &project.Currency, &project.IsActive,
		&project.IsLocked, &project.LockedAt, &project.LockReason, &project.LockedBy, &project.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	query := `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar, t.payee_payer,
		       fc.name, fc.type, fc.currency, t.transaction_date, t.created_at
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

	transactions := make([]models.ProjectTransactionDetail, 0)
	var totalIncome, totalExpenses float64

	for rows.Next() {
		var ptd models.ProjectTransactionDetail
		err := rows.Scan(
			&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
			&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
			&ptd.ContainerName, &ptd.ContainerType, &ptd.ContainerCurrency, &ptd.TransactionDate, &ptd.CreatedAt,
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

func (s *ReportService) GetDailyPosition(ctx context.Context) (*models.DailyPosition, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar,
		       account_number, iban, location, location_ar, is_active, created_at
		FROM financial_containers
		WHERE is_active = true
		ORDER BY type, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cashContainers := make([]models.FinancialContainer, 0)
	bankContainers := make([]models.FinancialContainer, 0)

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

		if c.Type == "cash" {
			cashContainers = append(cashContainers, c)
		} else if c.Type == "bank" {
			bankContainers = append(bankContainers, c)
		}
	}

	return &models.DailyPosition{
		CashContainers: cashContainers,
		BankContainers: bankContainers,
		GeneratedAt:    time.Now(),
	}, nil
}

func (s *ReportService) GetProjectSummary(ctx context.Context) ([]models.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active,
		       is_locked, locked_at, lock_reason, locked_by, created_at
		FROM projects
		WHERE is_active = true
		ORDER BY code, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]models.Project, 0)
	for rows.Next() {
		var p models.Project
		err := rows.Scan(
			&p.ID, &p.Code, &p.Name, &p.NameAr, &p.Description,
			&p.DonorName, &p.DonorNameAr, &p.StartDate, &p.EndDate,
			&p.Budget, &p.Currency, &p.IsActive,
			&p.IsLocked, &p.LockedAt, &p.LockReason, &p.LockedBy, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}

func (s *ReportService) GetInvoices(ctx context.Context, startDate, endDate *time.Time, projectID, containerID, transactionNumber string) ([]models.ProjectTransactionDetail, error) {
	query := `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar, t.payee_payer,
		       COALESCE(p.code, ''), COALESCE(p.name, ''),
		       fc.name, fc.type, fc.currency, t.transaction_date, t.created_at
		FROM transactions t
		LEFT JOIN projects p ON t.project_id = p.id
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.status = 'completed'
	`
	args := []interface{}{}
	argCount := 1

	if startDate != nil {
		query += ` AND t.transaction_date >= $` + string(rune(argCount+'0'))
		args = append(args, startDate)
		argCount++
	}
	if endDate != nil {
		query += ` AND t.transaction_date <= $` + string(rune(argCount+'0'))
		args = append(args, endDate)
		argCount++
	}
	if projectID != "" {
		query += ` AND t.project_id = $` + string(rune(argCount+'0'))
		args = append(args, projectID)
		argCount++
	}
	if containerID != "" {
		query += ` AND t.container_id = $` + string(rune(argCount+'0'))
		args = append(args, containerID)
		argCount++
	}
	if transactionNumber != "" {
		// Use ILIKE for partial match or exact match? Transaction numbers are usually exact. Let's use ILIKE for user friendliness.
		query += ` AND t.transaction_number ILIKE $` + string(rune(argCount+'0'))
		args = append(args, "%"+transactionNumber+"%")
		argCount++
	}

	query += ` ORDER BY t.transaction_number ASC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := make([]models.ProjectTransactionDetail, 0)
	for rows.Next() {
		var ptd models.ProjectTransactionDetail
		var containerCurrency models.CurrencyCode
		err := rows.Scan(
			&ptd.ID, &ptd.TransactionNumber, &ptd.Type, &ptd.Amount, &ptd.Currency,
			&ptd.Description, &ptd.DescriptionAr, &ptd.PayeePayer,
			&ptd.ProjectCode, &ptd.ProjectName,
			&ptd.ContainerName, &ptd.ContainerType, &containerCurrency, &ptd.TransactionDate, &ptd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		ptd.Currency = containerCurrency // Or transaction currency is usually same, but struct has one field.
		invoices = append(invoices, ptd)
	}

	return invoices, nil
}

// GetEmployeeStatement generates a detailed financial statement for an employee
func (s *ReportService) GetEmployeeStatement(ctx context.Context, employeeID string, startDate, endDate *time.Time) (*models.EmployeeStatement, error) {
	// Get employee details
	var employee models.Employee
	err := s.db.QueryRowContext(ctx, `
		SELECT id, employee_number, full_name, full_name_ar, position, position_ar,
		       department, department_ar, phone, email, salary, base_salary, 
		       monthly_budget, salary_currency, join_date, is_active, created_at
		FROM employees
		WHERE id = $1
	`, employeeID).Scan(
		&employee.ID, &employee.EmployeeNumber, &employee.FullName, &employee.FullNameAr,
		&employee.Position, &employee.PositionAr, &employee.Department, &employee.DepartmentAr,
		&employee.Phone, &employee.Email, &employee.Salary, &employee.BaseSalary,
		&employee.MonthlyBudget, &employee.SalaryCurrency, &employee.JoinDate, &employee.IsActive, &employee.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Get transactions for this employee
	query := `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar,
		       COALESCE(p.code, ''), COALESCE(p.name, ''),
		       fc.name, fc.type, t.transaction_date, t.created_at
		FROM transactions t
		LEFT JOIN projects p ON t.project_id = p.id
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.employee_id = $1 AND t.status = 'completed'
	`
	args := []interface{}{employeeID}
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

	transactions := make([]models.EmployeeTransactionDetail, 0)
	var totalSalary, totalCustody, totalBonus float64

	for rows.Next() {
		var etd models.EmployeeTransactionDetail
		err := rows.Scan(
			&etd.ID, &etd.TransactionNumber, &etd.Type, &etd.Amount, &etd.Currency,
			&etd.Description, &etd.DescriptionAr,
			&etd.ProjectCode, &etd.ProjectName,
			&etd.ContainerName, &etd.ContainerType, &etd.TransactionDate, &etd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Calculate totals by transaction type
		switch etd.Type {
		case models.TransactionDisbursementSalary:
			totalSalary += etd.Amount
		case models.TransactionDisbursementCustody:
			totalCustody += etd.Amount
		case models.TransactionDisbursementBonus:
			totalBonus += etd.Amount
		}

		transactions = append(transactions, etd)
	}

	// Get deductions for this employee
	deductionQuery := `
		SELECT id, employee_id, month_year, amount, currency, reason, reason_ar, created_at
		FROM salary_deductions
		WHERE employee_id = $1
	`
	deductionArgs := []interface{}{employeeID}

	if startDate != nil {
		deductionQuery += ` AND month_year >= $2`
		deductionArgs = append(deductionArgs, startDate)
	}
	if endDate != nil {
		if len(deductionArgs) == 1 {
			deductionQuery += ` AND month_year <= $2`
		} else {
			deductionQuery += ` AND month_year <= $3`
		}
		deductionArgs = append(deductionArgs, endDate)
	}

	deductionQuery += ` ORDER BY month_year DESC`

	deductionRows, err := s.db.QueryContext(ctx, deductionQuery, deductionArgs...)
	if err != nil {
		return nil, err
	}
	defer deductionRows.Close()

	deductions := make([]models.SalaryDeduction, 0)
	var totalDeductions float64

	for deductionRows.Next() {
		var d models.SalaryDeduction
		err := deductionRows.Scan(
			&d.ID, &d.EmployeeID, &d.MonthYear, &d.Amount, &d.Currency,
			&d.Reason, &d.ReasonAr, &d.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		totalDeductions += d.Amount
		deductions = append(deductions, d)
	}

	statement := &models.EmployeeStatement{
		Employee:     employee,
		Transactions: transactions,
		Deductions:   deductions,
		Summary: models.EmployeeStatementSummary{
			TotalSalary:      totalSalary,
			TotalCustody:     totalCustody,
			TotalBonus:       totalBonus,
			TotalDeductions:  totalDeductions,
			NetPayments:      totalSalary + totalBonus - totalDeductions,
			TransactionCount: len(transactions),
		},
		GeneratedAt: time.Now(),
	}

	return statement, nil
}

// GetMemberStatement generates a detailed financial statement for a member
func (s *ReportService) GetMemberStatement(ctx context.Context, memberID string, startDate, endDate *time.Time) (*models.MemberStatement, error) {
	// Get member details using the view for computed status
	var member models.Member
	err := s.db.QueryRowContext(ctx, `
		SELECT id, membership_number, full_name, full_name_ar, member_type, phone, email,
		       address, address_ar, join_date, original_status, valid_until, last_renewal_date,
		       subscription_amount, subscription_currency, computed_status, expiring_soon,
		       days_remaining, is_active, created_at
		FROM vw_members_with_status
		WHERE id = $1
	`, memberID).Scan(
		&member.ID, &member.MembershipNumber, &member.FullName, &member.FullNameAr,
		&member.MemberType, &member.Phone, &member.Email, &member.Address, &member.AddressAr,
		&member.JoinDate, &member.Status, &member.ValidUntil, &member.LastRenewalDate,
		&member.SubscriptionAmount, &member.SubscriptionCurrency, &member.ComputedStatus,
		&member.ExpiringSoon, &member.DaysRemaining, &member.IsActive, &member.CreatedAt,
	)
	if err != nil {
		// Fallback to regular members table if view doesn't exist
		err = s.db.QueryRowContext(ctx, `
			SELECT id, membership_number, full_name, full_name_ar, member_type, phone, email,
			       address, address_ar, join_date, status, valid_until, last_renewal_date,
			       subscription_amount, subscription_currency, is_active, created_at
			FROM members
			WHERE id = $1
		`, memberID).Scan(
			&member.ID, &member.MembershipNumber, &member.FullName, &member.FullNameAr,
			&member.MemberType, &member.Phone, &member.Email, &member.Address, &member.AddressAr,
			&member.JoinDate, &member.Status, &member.ValidUntil, &member.LastRenewalDate,
			&member.SubscriptionAmount, &member.SubscriptionCurrency, &member.IsActive, &member.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
	}

	// Get subscriptions for this member
	subRows, err := s.db.QueryContext(ctx, `
		SELECT id, member_id, transaction_id, amount, currency,
		       payment_date, valid_from, valid_until, notes, notes_ar, created_at
		FROM member_subscriptions
		WHERE member_id = $1
		ORDER BY payment_date DESC
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer subRows.Close()

	subscriptions := make([]models.MemberSubscription, 0)
	for subRows.Next() {
		var sub models.MemberSubscription
		err := subRows.Scan(
			&sub.ID, &sub.MemberID, &sub.TransactionID, &sub.Amount, &sub.Currency,
			&sub.PaymentDate, &sub.ValidFrom, &sub.ValidUntil, &sub.Notes, &sub.NotesAr, &sub.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		subscriptions = append(subscriptions, sub)
	}

	// Get transactions for this member
	query := `
		SELECT t.id, t.transaction_number, t.type, t.amount, t.currency,
		       t.description, t.description_ar,
		       COALESCE(p.code, ''), COALESCE(p.name, ''),
		       fc.name, t.transaction_date, t.created_at
		FROM transactions t
		LEFT JOIN projects p ON t.project_id = p.id
		JOIN financial_containers fc ON t.container_id = fc.id
		WHERE t.member_id = $1 AND t.status = 'completed'
	`
	args := []interface{}{memberID}
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

	transactions := make([]models.MemberTransactionDetail, 0)
	var totalDeposits float64

	for rows.Next() {
		var mtd models.MemberTransactionDetail
		err := rows.Scan(
			&mtd.ID, &mtd.TransactionNumber, &mtd.Type, &mtd.Amount, &mtd.Currency,
			&mtd.Description, &mtd.DescriptionAr,
			&mtd.ProjectCode, &mtd.ProjectName,
			&mtd.ContainerName, &mtd.TransactionDate, &mtd.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Sum up deposits (income transactions)
		if mtd.Type == models.TransactionIncomeGrant || mtd.Type == models.TransactionIncomeDonation {
			totalDeposits += mtd.Amount
		}

		transactions = append(transactions, mtd)
	}

	// Calculate membership years
	membershipYears := 0
	if member.JoinDate != nil {
		membershipYears = int(time.Since(*member.JoinDate).Hours() / 24 / 365)
	}

	statement := &models.MemberStatement{
		Member:        member,
		Subscriptions: subscriptions,
		Transactions:  transactions,
		Summary: models.MemberStatementSummary{
			TotalDeposits:     totalDeposits,
			SubscriptionsPaid: len(subscriptions),
			MembershipYears:   membershipYears,
			TransactionCount:  len(transactions),
		},
		GeneratedAt: time.Now(),
	}

	return statement, nil
}
