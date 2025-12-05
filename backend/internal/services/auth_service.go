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

// AuthService handles authentication operations with enhanced security
type AuthService struct {
	db           *database.DB
	cfg          *config.Config
	auditService *security.AuditService
	otpService   *security.OTPService
	emailService *EmailService
}

// NewAuthService creates a new authentication service
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
	// ErrInvalidCredentials indicates authentication failure
	ErrInvalidCredentials = errors.New("invalid username or password")
	// ErrUserNotActive indicates the user account is disabled
	ErrUserNotActive = errors.New("user account is not active")
	// ErrEmailDomainNotAllowed indicates email is not from allowed domain
	ErrEmailDomainNotAllowed = errors.New("only @lco.org.ly email addresses are allowed")
	// ErrWeakPassword indicates the password doesn't meet requirements
	ErrWeakPassword = errors.New("password must be at least 8 characters")
	// Err2FARequired indicates 2FA verification is needed
	Err2FARequired = errors.New("2FA verification required")
)

// LoginResult represents the result of a login attempt
type LoginResult struct {
	User          *models.User
	Requires2FA   bool
	PendingUserID uuid.UUID
	Message       string
}

// Login authenticates user credentials (Step 1 of 2FA flow)
// If 2FA is enabled, returns Requires2FA=true and user must call Verify2FA
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	var user models.User
	var passwordHash string

	// Fetch user with all required fields
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

	// Verify password using Argon2id
	valid, err := security.VerifyPassword(password, passwordHash)
	if err != nil || !valid {
		return nil, ErrInvalidCredentials
	}

	// Check if 2FA is enabled
	if user.TwoFactorEnabled {
		// Generate and send OTP
		code, err := s.otpService.GenerateOTP(user.ID)
		if err != nil {
			return nil, err
		}

		// Send OTP via email
		err = s.emailService.SendOTPEmail(ctx, user.Email, user.FullName, code)
		if err != nil {
			return nil, errors.New("failed to send verification code")
		}

		return &LoginResult{
			User:          nil, // Don't return user yet
			Requires2FA:   true,
			PendingUserID: user.ID,
			Message:       "Verification code sent to your email",
		}, nil
	}

	// No 2FA - complete login immediately
	// Update last login timestamp
	_, _ = s.db.ExecContext(ctx, `
		UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1
	`, user.ID)

	return &LoginResult{
		User:        &user,
		Requires2FA: false,
	}, nil
}

// Verify2FA completes the login after OTP verification (Step 2 of 2FA flow)
func (s *AuthService) Verify2FA(ctx context.Context, userID uuid.UUID, code string) (*models.User, error) {
	// Verify OTP
	err := s.otpService.VerifyOTP(userID, code)
	if err != nil {
		return nil, err
	}

	// Get user data
	user, err := s.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Update last login
	_, _ = s.db.ExecContext(ctx, `
		UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = $1
	`, userID)

	return user, nil
}

// ResendOTP generates and sends a new OTP code
func (s *AuthService) ResendOTP(ctx context.Context, userID uuid.UUID) error {
	// Get user email
	var email, fullName string
	err := s.db.QueryRowContext(ctx, `
		SELECT email, full_name FROM users WHERE id = $1
	`, userID).Scan(&email, &fullName)
	if err != nil {
		return err
	}

	// Generate new OTP
	code, err := s.otpService.GenerateOTP(userID)
	if err != nil {
		return err
	}

	// Send OTP via email
	return s.emailService.SendOTPEmail(ctx, email, fullName, code)
}

// CreateUser creates a new user with Argon2id password hashing
// Validates email domain restriction to @lco.org.ly
func (s *AuthService) CreateUser(ctx context.Context, username, email, password, fullName string, role models.UserRole, createdBy uuid.UUID) (*models.User, error) {
	// Validate email domain
	if err := security.ValidateEmail(email); err != nil {
		if err == security.ErrUnauthorizedDomain {
			return nil, ErrEmailDomainNotAllowed
		}
		return nil, err
	}

	// Validate password strength
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	// Hash password using Argon2id with secure defaults
	hashedPassword, err := security.HashPassword(password, nil)
	if err != nil {
		return nil, err
	}

	var user models.User
	// Use NULL for created_by during self-registration
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

// GetUserByID retrieves user by ID
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

// RevokeAllSessions increments token version to invalidate all existing sessions
// This is the "Kill Switch" functionality
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

// ChangePassword updates user password and invalidates all sessions
func (s *AuthService) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error {
	// Validate new password strength
	if len(newPassword) < 8 {
		return ErrWeakPassword
	}

	// Get current password hash
	var currentHash string
	err := s.db.QueryRowContext(ctx, `
		SELECT password_hash FROM users WHERE id = $1
	`, userID).Scan(&currentHash)

	if err != nil {
		return err
	}

	// Verify old password
	valid, err := security.VerifyPassword(oldPassword, currentHash)
	if err != nil || !valid {
		return ErrInvalidCredentials
	}

	// Hash new password
	newHash, err := security.HashPassword(newPassword, nil)
	if err != nil {
		return err
	}

	// Update password and increment token version to invalidate sessions
	_, err = s.db.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = $1, 
		    token_version = token_version + 1, 
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, newHash, userID)

	return err
}

// Toggle2FA enables or disables 2FA for a user
func (s *AuthService) Toggle2FA(ctx context.Context, userID uuid.UUID, enabled bool) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE users SET two_factor_enabled = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`, enabled, userID)
	return err
}

// ValidateTokenVersion checks if the token version matches the database
// Used by AuthMiddleware for zero-trust validation
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

// GetAuditService returns the audit service for logging
func (s *AuthService) GetAuditService() *security.AuditService {
	return s.auditService
}
