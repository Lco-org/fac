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
