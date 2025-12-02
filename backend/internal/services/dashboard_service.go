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

// GetWidgetConfigs retrieves all dashboard widget configurations
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

	var widgets []models.DashboardWidget
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

	return widgets, nil
}

// GetVisibleWidgets retrieves widgets visible to a specific role
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

	var widgets []models.DashboardWidget
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

	return widgets, nil
}

// UpdateWidgetConfig updates a widget configuration
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
