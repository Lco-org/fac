package services

import (
	"context"
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db  *database.DB
	cfg *config.Config
}

func NewAuthService(db *database.DB, cfg *config.Config) *AuthService {
	return &AuthService{db: db, cfg: cfg}
}

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUserNotActive      = errors.New("user account is not active")
)

// Login authenticates user and returns user data
func (s *AuthService) Login(ctx context.Context, username, password string) (*models.User, error) {
	var user models.User

	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, full_name, full_name_ar, role, is_active, created_at
		FROM users
		WHERE username = $1 OR email = $1
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.FullNameAr,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if !user.IsActive {
		return nil, ErrUserNotActive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login
	_, _ = s.db.ExecContext(ctx, `
		UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1
	`, user.ID)

	return &user, nil
}

// CreateUser creates a new user
func (s *AuthService) CreateUser(ctx context.Context, username, email, password, fullName string, role models.UserRole, createdBy uuid.UUID) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), s.cfg.Security.BcryptCost)
	if err != nil {
		return nil, err
	}

	var user models.User
	// Use NULL for created_by during self-registration
	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, full_name, role, created_by)
		VALUES ($1, $2, $3, $4, $5, NULL)
		RETURNING id, username, email, full_name, role, is_active, created_at
	`, username, email, string(hashedPassword), fullName, role).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
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
		return nil, err
	}

	return &user, nil
}
