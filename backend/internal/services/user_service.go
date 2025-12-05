package services

import (
	"context"
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"cso-erp/internal/security"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

// UserService handles user management operations
type UserService struct {
	db  *database.DB
	cfg *config.Config
}

// NewUserService creates a new user service
func NewUserService(db *database.DB, cfg *config.Config) *UserService {
	return &UserService{db: db, cfg: cfg}
}

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUsernameExists   = errors.New("username already exists")
	ErrEmailExists      = errors.New("email already exists")
	ErrCannotDeleteSelf = errors.New("cannot delete your own account")
)

// ListUsers returns all users
func (s *UserService) ListUsers(ctx context.Context) ([]models.User, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, username, email, full_name, full_name_ar, role, is_active, last_login, created_at
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users = []models.User{}
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Email,
			&user.FullName,
			&user.FullNameAr,
			&user.Role,
			&user.IsActive,
			&user.LastLogin,
			&user.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, email, full_name, full_name_ar, role, is_active, last_login, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.FullNameAr,
		&user.Role,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}

// CreateUser creates a new user with Argon2id password hashing
func (s *UserService) CreateUser(ctx context.Context, req *models.CreateUserRequest, createdBy uuid.UUID) (*models.User, error) {
	// Validate email domain
	if err := security.ValidateEmail(req.Email); err != nil {
		if err == security.ErrUnauthorizedDomain {
			return nil, errors.New("only @lco.org.ly email addresses are allowed")
		}
		return nil, err
	}

	// Check if username exists
	var exists bool
	err := s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", req.Username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameExists
	}

	// Check if email exists
	email := security.NormalizeEmail(req.Email)
	err = s.db.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Validate password strength
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}

	// Hash password using Argon2id
	hashedPassword, err := security.HashPassword(req.Password, nil)
	if err != nil {
		return nil, err
	}

	var user models.User
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, full_name, full_name_ar, role, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, username, email, full_name, full_name_ar, role, is_active, created_at
	`, req.Username, email, hashedPassword, req.FullName, req.FullNameAr, req.Role, createdBy).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.FullNameAr,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, id uuid.UUID, req *models.UpdateUserRequest) (*models.User, error) {
	// Build dynamic update query
	query := "UPDATE users SET "
	args := []interface{}{}
	argCount := 1

	if req.Email != nil {
		// Validate email domain
		if err := security.ValidateEmail(*req.Email); err != nil {
			if err == security.ErrUnauthorizedDomain {
				return nil, errors.New("only @lco.org.ly email addresses are allowed")
			}
			return nil, err
		}
		email := security.NormalizeEmail(*req.Email)
		query += fmt.Sprintf("email = $%d, ", argCount)
		args = append(args, email)
		argCount++
	}
	if req.FullName != nil {
		query += fmt.Sprintf("full_name = $%d, ", argCount)
		args = append(args, *req.FullName)
		argCount++
	}
	if req.FullNameAr != nil {
		query += fmt.Sprintf("full_name_ar = $%d, ", argCount)
		args = append(args, *req.FullNameAr)
		argCount++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf("is_active = $%d, ", argCount)
		args = append(args, *req.IsActive)
		argCount++
	}

	// Remove trailing comma
	query = query[:len(query)-2]
	query += fmt.Sprintf(" WHERE id = $%d RETURNING id, username, email, full_name, full_name_ar, role, is_active, created_at", argCount)
	args = append(args, id)

	var user models.User
	err := s.db.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.FullNameAr,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// DeleteUser deletes a user permanently
func (s *UserService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM users WHERE id = $1
	`, id)

	if err != nil {
		// Check for foreign key constraint violation
		return errors.New("cannot delete user: associated records exist (try deactivating instead)")
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateUserRole updates user role
func (s *UserService) UpdateUserRole(ctx context.Context, id uuid.UUID, role models.UserRole) (*models.User, error) {
	var user models.User
	err := s.db.QueryRowContext(ctx, `
		UPDATE users SET role = $1
		WHERE id = $2
		RETURNING id, username, email, full_name, full_name_ar, role, is_active, created_at
	`, role, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.FullNameAr,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, ErrUserNotFound
	}

	return &user, nil
}
