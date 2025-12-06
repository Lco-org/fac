package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"errors"

	"github.com/google/uuid"
)

type EmployeeService struct {
	db *database.DB
}

func NewEmployeeService(db *database.DB) *EmployeeService {
	return &EmployeeService{db: db}
}

func (s *EmployeeService) ListEmployees(ctx context.Context) ([]models.Employee, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, employee_number, full_name, full_name_ar, position, position_ar,
		       department, department_ar, phone, email, salary, join_date, is_active, created_at
		FROM employees
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees = []models.Employee{}
	for rows.Next() {
		var e models.Employee
		err := rows.Scan(
			&e.ID, &e.EmployeeNumber, &e.FullName, &e.FullNameAr, &e.Position, &e.PositionAr,
			&e.Department, &e.DepartmentAr, &e.Phone, &e.Email, &e.Salary, &e.JoinDate,
			&e.IsActive, &e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, e)
	}

	return employees, nil
}

func (s *EmployeeService) GetEmployeeByID(ctx context.Context, id uuid.UUID) (*models.Employee, error) {
	var e models.Employee
	err := s.db.QueryRowContext(ctx, `
		SELECT id, employee_number, full_name, full_name_ar, position, position_ar,
		       department, department_ar, phone, email, salary, join_date, is_active, created_at
		FROM employees
		WHERE id = $1
	`, id).Scan(
		&e.ID, &e.EmployeeNumber, &e.FullName, &e.FullNameAr, &e.Position, &e.PositionAr,
		&e.Department, &e.DepartmentAr, &e.Phone, &e.Email, &e.Salary, &e.JoinDate,
		&e.IsActive, &e.CreatedAt,
	)

	if err != nil {
		return nil, ErrEmployeeNotFound
	}

	return &e, nil
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, req *models.CreateEmployeeRequest, createdBy uuid.UUID) (*models.Employee, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM employees WHERE employee_number = $1)", req.EmployeeNumber).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmployeeNumberExists
	}

	var e models.Employee
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO employees (employee_number, full_name, full_name_ar, position, position_ar,
		                       department, department_ar, phone, email, salary, join_date, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, employee_number, full_name, full_name_ar, position, position_ar,
		          department, department_ar, phone, email, salary, join_date, is_active, created_at
	`, req.EmployeeNumber, req.FullName, req.FullNameAr, req.Position, req.PositionAr,
		req.Department, req.DepartmentAr, req.Phone, req.Email, req.Salary, req.JoinDate, createdBy).Scan(
		&e.ID, &e.EmployeeNumber, &e.FullName, &e.FullNameAr, &e.Position, &e.PositionAr,
		&e.Department, &e.DepartmentAr, &e.Phone, &e.Email, &e.Salary, &e.JoinDate,
		&e.IsActive, &e.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &e, nil
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, id uuid.UUID, req *models.UpdateEmployeeRequest) (*models.Employee, error) {
	var e models.Employee
	err := s.db.QueryRowContext(ctx, `
		UPDATE employees SET
			full_name = COALESCE($1, full_name),
			full_name_ar = COALESCE($2, full_name_ar),
			position = COALESCE($3, position),
			position_ar = COALESCE($4, position_ar),
			department = COALESCE($5, department),
			department_ar = COALESCE($6, department_ar),
			phone = COALESCE($7, phone),
			email = COALESCE($8, email),
			salary = COALESCE($9, salary),
			join_date = COALESCE($10, join_date),
			is_active = COALESCE($11, is_active),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $12
		RETURNING id, employee_number, full_name, full_name_ar, position, position_ar,
		          department, department_ar, phone, email, salary, join_date, is_active, created_at
	`, req.FullName, req.FullNameAr, req.Position, req.PositionAr, req.Department, req.DepartmentAr,
		req.Phone, req.Email, req.Salary, req.JoinDate, req.IsActive, id).Scan(
		&e.ID, &e.EmployeeNumber, &e.FullName, &e.FullNameAr, &e.Position, &e.PositionAr,
		&e.Department, &e.DepartmentAr, &e.Phone, &e.Email, &e.Salary, &e.JoinDate,
		&e.IsActive, &e.CreatedAt,
	)

	if err != nil {
		return nil, ErrEmployeeNotFound
	}

	return &e, nil
}

func (s *EmployeeService) DeleteEmployee(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM employees WHERE id = $1
	`, id)

	if err != nil {
		return errors.New("cannot delete employee: associated records exist")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrEmployeeNotFound
	}

	return nil
}
