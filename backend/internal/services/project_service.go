package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"errors"

	"github.com/google/uuid"
)

type ProjectService struct {
	db *database.DB
}

func NewProjectService(db *database.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]models.Project, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active,
		       is_locked, locked_at, lock_reason, locked_by, created_at
		FROM projects
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects = []models.Project{}
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

func (s *ProjectService) GetProjectByID(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	var p models.Project
	err := s.db.QueryRowContext(ctx, `
		SELECT id, code, name, name_ar, description, donor_name, donor_name_ar,
		       start_date, end_date, budget, currency, is_active,
		       is_locked, locked_at, lock_reason, locked_by, created_at
		FROM projects
		WHERE id = $1
	`, id).Scan(
		&p.ID, &p.Code, &p.Name, &p.NameAr, &p.Description,
		&p.DonorName, &p.DonorNameAr, &p.StartDate, &p.EndDate,
		&p.Budget, &p.Currency, &p.IsActive,
		&p.IsLocked, &p.LockedAt, &p.LockReason, &p.LockedBy, &p.CreatedAt,
	)

	if err != nil {
		return nil, ErrProjectNotFound
	}

	return &p, nil
}

func (s *ProjectService) CreateProject(ctx context.Context, req *models.CreateProjectRequest, createdBy uuid.UUID) (*models.Project, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM projects WHERE code = $1)", req.Code).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrProjectCodeExists
	}

	var p models.Project
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO projects (code, name, name_ar, description, donor_name, donor_name_ar,
		                      start_date, end_date, budget, currency, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, code, name, name_ar, description, donor_name, donor_name_ar,
		          start_date, end_date, budget, currency, is_active,
		          is_locked, locked_at, lock_reason, locked_by, created_at
	`, req.Code, req.Name, req.NameAr, req.Description, req.DonorName, req.DonorNameAr,
		req.StartDate, req.EndDate, req.Budget, req.Currency, createdBy).Scan(
		&p.ID, &p.Code, &p.Name, &p.NameAr, &p.Description,
		&p.DonorName, &p.DonorNameAr, &p.StartDate, &p.EndDate,
		&p.Budget, &p.Currency, &p.IsActive,
		&p.IsLocked, &p.LockedAt, &p.LockReason, &p.LockedBy, &p.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, id uuid.UUID, req *models.UpdateProjectRequest) (*models.Project, error) {
	var p models.Project
	err := s.db.QueryRowContext(ctx, `
		UPDATE projects SET
			name = COALESCE($1, name),
			name_ar = COALESCE($2, name_ar),
			description = COALESCE($3, description),
			donor_name = COALESCE($4, donor_name),
			donor_name_ar = COALESCE($5, donor_name_ar),
			start_date = COALESCE($6, start_date),
			end_date = COALESCE($7, end_date),
			budget = COALESCE($8, budget),
			currency = COALESCE($9, currency),
			is_active = COALESCE($10, is_active),
			is_locked = COALESCE($11, is_locked),
			locked_at = CASE WHEN $11 = true AND is_locked = false THEN CURRENT_TIMESTAMP ELSE locked_at END,
			lock_reason = COALESCE($12, lock_reason),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $13
		RETURNING id, code, name, name_ar, description, donor_name, donor_name_ar,
		          start_date, end_date, budget, currency, is_active,
		          is_locked, locked_at, lock_reason, locked_by, created_at
	`, req.Name, req.NameAr, req.Description, req.DonorName, req.DonorNameAr,
		req.StartDate, req.EndDate, req.Budget, req.Currency, req.IsActive,
		req.IsLocked, req.LockReason, id).Scan(
		&p.ID, &p.Code, &p.Name, &p.NameAr, &p.Description,
		&p.DonorName, &p.DonorNameAr, &p.StartDate, &p.EndDate,
		&p.Budget, &p.Currency, &p.IsActive,
		&p.IsLocked, &p.LockedAt, &p.LockReason, &p.LockedBy, &p.CreatedAt,
	)

	if err != nil {
		return nil, ErrProjectNotFound
	}

	return &p, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM projects WHERE id = $1
	`, id)

	if err != nil {
		return errors.New("cannot delete project: associated transactions exist")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrProjectNotFound
	}

	return nil
}
