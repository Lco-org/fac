package services

import (
	"context"
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"cso-erp/internal/security"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type AuthService struct {
	db           *database.DB
	cfg          *config.Config
	auditService *security.AuditService
	otpService   *security.OTPService
	emailService *EmailService
}

func NewAuthService(db *database.DB, cfg *config.Config, emailService *EmailService) *AuthService {
	return &AuthService{
		db:           db,
		cfg:          cfg,
		auditService: security.NewAuditService(db.DB),
		otpService:   security.NewOTPService(db.DB),
		emailService: emailService,
	}
}

var (
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrUserNotActive         = errors.New("user account is not active")
	ErrEmailDomainNotAllowed = errors.New("only @lco.org.ly email addresses are allowed")
	ErrWeakPassword          = errors.New("password must be at least 8 characters")
	Err2FARequired           = errors.New("2FA verification required")
)

type LoginResult struct {
	User          *models.User
	Requires2FA   bool
	PendingUserID uuid.UUID
	Message       string
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var user models.User
	var passwordHash string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, email, password_hash, full_name, full_name_ar, 
		       role, token_version, two_factor_enabled, is_active, created_at
		FROM users
		WHERE username = $1 OR email = $1
	`, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&passwordHash,
		&user.FullName,
		&user.FullNameAr,
		&user.Role,
		&user.TokenVersion,
		&user.TwoFactorEnabled,
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

	valid, err := security.VerifyPassword(password, passwordHash)
	if err != nil || !valid {
		return nil, ErrInvalidCredentials
	}

	if user.TwoFactorEnabled {
		code, err := s.otpService.GenerateOTP(user.ID)
		if err != nil {
			return nil, err
		}

		err = s.emailService.SendOTPEmail(ctx, user.Email, user.FullName, code)
		if err != nil {
			return nil, errors.New("failed to send verification code")
		}

		return &LoginResult{
			User:          nil,
			Requires2FA:   true,
			PendingUserID: user.ID,
			Message:       "Verification code sent to your email",
		}, nil
	}

	_, _ = s.db.ExecContext(ctx, `
		UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1
	`, user.ID)

	return &LoginResult{
		User:        &user,
		Requires2FA: false,
	}, nil
}

func (s *AuthService) Verify2FA(ctx context.Context, userID uuid.UUID, code string) (*models.User, error) {
	err := s.otpService.VerifyOTP(userID, code)
	if err != nil {
		return nil, err
	}

	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	_, _ = s.db.ExecContext(ctx, `
		UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1
	`, userID)

	return user, nil
}

func (s *AuthService) ResendOTP(ctx context.Context, userID uuid.UUID) error {
	var email, fullName string
	err := s.db.QueryRowContext(ctx, `
		SELECT email, full_name FROM users WHERE id = $1
	`, userID).Scan(&email, &fullName)
	if err != nil {
		return err
	}

	code, err := s.otpService.GenerateOTP(userID)
	if err != nil {
		return err
	}

	return s.emailService.SendOTPEmail(ctx, email, fullName, code)
}

func (s *AuthService) CreateUser(ctx context.Context, username, email, password, fullName string, role models.UserRole, createdBy uuid.UUID) (*models.User, error) {
	if err := security.ValidateEmail(email); err != nil {
		if err == security.ErrUnauthorizedDomain {
			return nil, ErrEmailDomainNotAllowed
		}
		return nil, err
	}

	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	hashedPassword, err := security.HashPassword(password, nil)
	if err != nil {
		return nil, err
	}

	var user models.User
	var createdByPtr *uuid.UUID
	if createdBy != uuid.Nil {
		createdByPtr = &createdBy
	}

	err = s.db.QueryRowContext(ctx, `
		INSERT INTO users (username, email, password_hash, full_name, role, two_factor_enabled, created_by)
		VALUES ($1, $2, $3, $4, $5, true, $6)
		RETURNING id, username, email, full_name, role, token_version, two_factor_enabled, is_active, created_at
	`, username, security.NormalizeEmail(email), hashedPassword, fullName, role, createdByPtr).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.Role,
		&user.TokenVersion,
		&user.TwoFactorEnabled,
		&user.IsActive,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User

	err := s.db.QueryRowContext(ctx, `
		SELECT id, username, email, full_name, full_name_ar, role, 
		       token_version, two_factor_enabled, is_active, last_login, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.FullName,
		&user.FullNameAr,
		&user.Role,
		&user.TokenVersion,
		&user.TwoFactorEnabled,
		&user.IsActive,
		&user.LastLogin,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *AuthService) RevokeAllSessions(ctx context.Context, userID uuid.UUID) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE users 
		SET token_version = token_version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`, userID)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	var currentHash string
	err := s.db.QueryRowContext(ctx, `
		SELECT password_hash FROM users WHERE id = $1
	`, userID).Scan(&currentHash)

	if err != nil {
		return err
	}

	valid, err := security.VerifyPassword(oldPassword, currentHash)
	if err != nil || !valid {
		return ErrInvalidCredentials
	}

	newHash, err := security.HashPassword(newPassword, nil)
	if err != nil {
		return err
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = $1, 
		    token_version = token_version + 1, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, newHash, userID)

	return err
}

func (s *AuthService) Toggle2FA(ctx context.Context, userID uuid.UUID, enabled bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET two_factor_enabled = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, enabled, userID)
	return err
}

func (s *AuthService) ValidateTokenVersion(ctx context.Context, userID uuid.UUID, tokenVersion int) (bool, error) {
	var currentVersion int
	err := s.db.QueryRowContext(ctx, `
		SELECT token_version FROM users WHERE id = $1 AND is_active = true
	`, userID).Scan(&currentVersion)

	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return currentVersion == tokenVersion, nil
}

func (s *AuthService) GetAuditService() *security.AuditService {
	return s.auditService
}
