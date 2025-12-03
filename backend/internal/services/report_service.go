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

// GetDashboardSummary retrieves the dashboard summary with detailed transactions filtered by user role
func (s *ReportService) GetDashboardSummary(ctx context.Context, userRole string) (*models.DashboardSummary, error) {
	// Get visible widgets for this role
	dashService := NewDashboardService(s.db)
	visibleWidgets, err := dashService.GetVisibleWidgets(ctx, userRole)
	if err != nil {
		return nil, err
	}

	// DEBUG: Log visible widgets
	log.Printf("DEBUG: User role: %s, Visible widgets count: %d", userRole, len(visibleWidgets))
	for _, w := range visibleWidgets {
		log.Printf("DEBUG: Widget: %s, Enabled: %v, Roles: %v", w.WidgetKey, w.IsEnabled, w.AllowedRoles)
	}

	// Create a map for quick lookup
	widgetMap := make(map[string]bool)
	for _, w := range visibleWidgets {
		widgetMap[w.WidgetKey] = true
	}

	log.Printf("DEBUG: Widget map: %v", widgetMap)

	// Initialize slices with empty arrays to ensure JSON returns [] instead of null
	bankSummaries := make([]models.BankDashboardSummary, 0)
	projectSummaries := make([]models.ProjectDashboardSummary, 0)
	recentIncome := make([]models.ProjectTransactionDetail, 0)
	recentExpenses := make([]models.ProjectTransactionDetail, 0)

	// 1. Get all active bank containers (if banks widget is visible)
	if widgetMap["banks"] {
		log.Println("DEBUG: Fetching bank accounts...")
		rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar, 
		       account_number, iban, location, location_ar, is_active, created_at
		FROM financial_containers
		WHERE type = 'bank' AND is_active = true
		ORDER BY name
	`)
		if err != nil {
			log.Printf("ERROR: Failed to fetch bank accounts: %v", err)
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
				log.Printf("ERROR: Failed to scan bank account: %v", err)
				return nil, err
			}

			// 2. For each bank, get last 3 transactions
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
				log.Printf("ERROR: Failed to fetch transactions for bank %s: %v", c.ID, err)
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
					log.Printf("ERROR: Failed to scan transaction: %v", err)
					return nil, err
				}
				// Strict currency enforcement: Always use container currency for display in this context
				td.Currency = c.Currency
				transactions = append(transactions, td)
			}
			log.Printf("DEBUG: Bank %s has %d transactions", c.Name, len(transactions))

			bankSummaries = append(bankSummaries, models.BankDashboardSummary{
				Container:        c,
				LastTransactions: transactions,
			})
		}
		log.Printf("DEBUG: Fetched %d bank accounts", len(bankSummaries))
	}

	// 3. Get active projects with last 3 transactions (if projects widget is visible)
	if widgetMap["projects"] {
		log.Println("DEBUG: Fetching projects...")
		projectRows, err := s.db.QueryContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active, created_at
		FROM projects
		WHERE is_active = true
		ORDER BY name
	`)
		if err != nil {
			log.Printf("ERROR: Failed to fetch projects: %v", err)
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
				log.Printf("ERROR: Failed to scan project: %v", err)
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
				log.Printf("ERROR: Failed to fetch transactions for project %s: %v", p.ID, err)
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
					log.Printf("ERROR: Failed to scan project transaction: %v", err)
					return nil, err
				}
				// Strict currency enforcement: Always use project currency for display
				ptd.Currency = p.Currency
				transactions = append(transactions, ptd)
			}
			log.Printf("DEBUG: Project %s has %d transactions", p.Name, len(transactions))

			projectSummaries = append(projectSummaries, models.ProjectDashboardSummary{
				Project:          p,
				LastTransactions: transactions,
			})
		}
		log.Printf("DEBUG: Fetched %d projects", len(projectSummaries))
	}

	// 4. Get recent income (global) - if recent_income widget is visible
	if widgetMap["recent_income"] {
		log.Println("DEBUG: Fetching recent income...")
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
			log.Printf("ERROR: Failed to fetch recent income: %v", err)
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
				log.Printf("ERROR: Failed to scan income transaction: %v", err)
				return nil, err
			}
			// Strict currency enforcement
			ptd.Currency = containerCurrency
			recentIncome = append(recentIncome, ptd)
		}
		log.Printf("DEBUG: Fetched %d recent income transactions", len(recentIncome))
	}

	// 5. Get recent expenses (global) - if recent_expenses widget is visible
	if widgetMap["recent_expenses"] {
		log.Println("DEBUG: Fetching recent expenses...")
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
			log.Printf("ERROR: Failed to fetch recent expenses: %v", err)
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
				log.Printf("ERROR: Failed to scan expense transaction: %v", err)
				return nil, err
			}
			// Strict currency enforcement
			ptd.Currency = containerCurrency
			recentExpenses = append(recentExpenses, ptd)
		}
		log.Printf("DEBUG: Fetched %d recent expense transactions", len(recentExpenses))
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

// GetDailyPosition returns financial containers grouped by type (cash/bank)
func (s *ReportService) GetDailyPosition(ctx context.Context) (*models.DailyPosition, error) {
	log.Println("DEBUG: Fetching daily position...")

	// Get all active containers
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, name_ar, type, currency, balance, bank_name, bank_name_ar,
		       account_number, iban, location, location_ar, is_active, created_at
		FROM financial_containers
		WHERE is_active = true
		ORDER BY type, name
	`)
	if err != nil {
		log.Printf("ERROR: Failed to fetch containers: %v", err)
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
			log.Printf("ERROR: Failed to scan container: %v", err)
			return nil, err
		}

		if c.Type == "cash" {
			cashContainers = append(cashContainers, c)
		} else if c.Type == "bank" {
			bankContainers = append(bankContainers, c)
		}
	}

	log.Printf("DEBUG: Found %d cash containers and %d bank containers", len(cashContainers), len(bankContainers))

	return &models.DailyPosition{
		CashContainers: cashContainers,
		BankContainers: bankContainers,
		GeneratedAt:    time.Now(),
	}, nil
}

// GetProjectSummary returns a list of all active projects
func (s *ReportService) GetProjectSummary(ctx context.Context) ([]models.Project, error) {
	log.Println("DEBUG: Fetching project summary...")

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active, created_at
		FROM projects
		WHERE is_active = true
		ORDER BY code, name
	`)
	if err != nil {
		log.Printf("ERROR: Failed to fetch projects: %v", err)
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
			log.Printf("ERROR: Failed to scan project: %v", err)
			return nil, err
		}
		projects = append(projects, p)
	}

	log.Printf("DEBUG: Found %d active projects", len(projects))
	return projects, nil
}
