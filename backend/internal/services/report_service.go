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
		       start_date, end_date, budget, currency, is_active, created_at
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
				&p.Budget, &p.Currency, &p.IsActive, &p.CreatedAt,
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

	transactions := make([]models.ProjectTransactionDetail, 0)
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
		       start_date, end_date, budget, currency, is_active, created_at
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
			&p.Budget, &p.Currency, &p.IsActive, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}

	return projects, nil
}
