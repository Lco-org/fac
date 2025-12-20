package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"

	"github.com/lib/pq"
)

type DashboardService struct {
	db *database.DB
}

func NewDashboardService(db *database.DB) *DashboardService {
	return &DashboardService{db: db}
}

func (s *DashboardService) GetWidgetConfigs(ctx context.Context) ([]models.DashboardWidget, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, widget_key, widget_name, widget_name_en, is_enabled, 
		       allowed_roles, display_order, created_at, updated_at
		FROM dashboard_widgets
		ORDER BY display_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	widgets := []models.DashboardWidget{}
	for rows.Next() {
		var w models.DashboardWidget
		err := rows.Scan(
			&w.ID, &w.WidgetKey, &w.WidgetName, &w.WidgetNameEn,
			&w.IsEnabled, pq.Array(&w.AllowedRoles), &w.DisplayOrder,
			&w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		widgets = append(widgets, w)
	}

	// Auto-seed if empty
	if len(widgets) == 0 {
		return s.SeedDefaultWidgets(ctx)
	}

	return widgets, nil
}

func (s *DashboardService) SeedDefaultWidgets(ctx context.Context) ([]models.DashboardWidget, error) {
	defaultWidgets := []models.DashboardWidget{
		{
			WidgetKey:    "financial_summary",
			WidgetName:   "الملخص المالي",
			WidgetNameEn: "Financial Summary",
			IsEnabled:    true,
			AllowedRoles: []string{"admin", "finance", "auditor"},
			DisplayOrder: 1,
		},
		{
			WidgetKey:    "banks",
			WidgetName:   "الموقف المصرفي",
			WidgetNameEn: "Bank Accounts",
			IsEnabled:    true,
			AllowedRoles: []string{"admin", "finance", "auditor"},
			DisplayOrder: 2,
		},
		{
			WidgetKey:    "cash_accounts",
			WidgetName:   "الموقف النقدي",
			WidgetNameEn: "Cash Accounts",
			IsEnabled:    true,
			AllowedRoles: []string{"admin", "finance", "auditor"},
			DisplayOrder: 3,
		},
		{
			WidgetKey:    "projects",
			WidgetName:   "المشاريع",
			WidgetNameEn: "Projects",
			IsEnabled:    true,
			AllowedRoles: []string{"admin", "finance", "auditor"},
			DisplayOrder: 4,
		},
		{
			WidgetKey:    "recent_income",
			WidgetName:   "آخر المقبوضات",
			WidgetNameEn: "Recent Income",
			IsEnabled:    true,
			AllowedRoles: []string{"admin", "finance", "auditor"},
			DisplayOrder: 5,
		},
		{
			WidgetKey:    "recent_expenses",
			WidgetName:   "آخر المصروفات",
			WidgetNameEn: "Recent Expenses",
			IsEnabled:    true,
			AllowedRoles: []string{"admin", "finance", "auditor"},
			DisplayOrder: 6,
		},
	}

	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO dashboard_widgets (widget_key, widget_name, widget_name_en, is_enabled, allowed_roles, display_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, created_at, updated_at
	`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var seededWidgets []models.DashboardWidget

	for _, w := range defaultWidgets {
		err := stmt.QueryRowContext(ctx,
			w.WidgetKey, w.WidgetName, w.WidgetNameEn, w.IsEnabled, pq.Array(w.AllowedRoles), w.DisplayOrder,
		).Scan(&w.ID, &w.CreatedAt, &w.UpdatedAt)

		if err != nil {
			return nil, err
		}
		seededWidgets = append(seededWidgets, w)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return seededWidgets, nil

}

func (s *DashboardService) GetVisibleWidgets(ctx context.Context, userRole string) ([]models.DashboardWidget, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, widget_key, widget_name, widget_name_en, is_enabled, 
		       allowed_roles, display_order, created_at, updated_at
		FROM dashboard_widgets
		WHERE is_enabled = true AND $1 = ANY(allowed_roles)
		ORDER BY display_order ASC
	`, userRole)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	widgets := []models.DashboardWidget{}
	for rows.Next() {
		var w models.DashboardWidget
		err := rows.Scan(
			&w.ID, &w.WidgetKey, &w.WidgetName, &w.WidgetNameEn,
			&w.IsEnabled, pq.Array(&w.AllowedRoles), &w.DisplayOrder,
			&w.CreatedAt, &w.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		widgets = append(widgets, w)
	}

	// If no widgets found, check if table is empty
	if len(widgets) == 0 {
		var count int
		if err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dashboard_widgets").Scan(&count); err == nil && count == 0 {
			// Table is empty, seed defaults
			seeded, err := s.SeedDefaultWidgets(ctx)
			if err != nil {
				return nil, err
			}

			// Filter seeded widgets for the user role
			filtered := []models.DashboardWidget{}
			for _, w := range seeded {
				if w.IsEnabled && contains(w.AllowedRoles, userRole) {
					filtered = append(filtered, w)
				}
			}
			return filtered, nil
		}
	}

	return widgets, nil
}

// Helper to check slice containment
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (s *DashboardService) UpdateWidgetConfig(ctx context.Context, widgetID string, req models.UpdateWidgetConfigRequest) error {
	query := "UPDATE dashboard_widgets SET updated_at = CURRENT_TIMESTAMP"
	args := []interface{}{}
	argCount := 1

	if req.IsEnabled != nil {
		query += ", is_enabled = $" + string(rune(argCount+'0'))
		args = append(args, *req.IsEnabled)
		argCount++
	}

	if req.AllowedRoles != nil {
		query += ", allowed_roles = $" + string(rune(argCount+'0'))
		args = append(args, pq.Array(req.AllowedRoles))
		argCount++
	}

	if req.DisplayOrder != nil {
		query += ", display_order = $" + string(rune(argCount+'0'))
		args = append(args, *req.DisplayOrder)
		argCount++
	}

	query += " WHERE id = $" + string(rune(argCount+'0'))
	args = append(args, widgetID)

	result, err := s.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (s *DashboardService) BulkUpdateWidgetOrder(ctx context.Context, updates []models.WidgetOrderUpdate) error {
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		UPDATE dashboard_widgets 
		SET display_order = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE id = $2
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, update := range updates {
		_, err := stmt.ExecContext(ctx, update.DisplayOrder, update.ID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
