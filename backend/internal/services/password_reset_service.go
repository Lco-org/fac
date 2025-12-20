package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"cso-erp/internal/database"
	apperrors "cso-erp/internal/errors"
	"cso-erp/internal/security"

	"github.com/google/uuid"
)

// PasswordResetService handles password reset operations
type PasswordResetService struct {
	db           *database.DB
	emailService *EmailService
	tokenExpiry  time.Duration
}

// NewPasswordResetService creates a new password reset service
func NewPasswordResetService(db *database.DB, emailService *EmailService, tokenExpiry time.Duration) *PasswordResetService {
	return &PasswordResetService{
		db:           db,
		emailService: emailService,
		tokenExpiry:  tokenExpiry,
	}
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
	UsedAt    *time.Time
	IP        string
}

// RequestPasswordReset initiates password reset flow
func (s *PasswordResetService) RequestPasswordReset(ctx context.Context, email, ip string) error {
	// Find user by email
	var userID uuid.UUID
	var fullName string
	err := s.db.QueryRowContext(ctx, `
		SELECT id, full_name FROM users WHERE email = $1 AND is_active = true
	`, email).Scan(&userID, &fullName)

	if err != nil {
		// Don't reveal if email exists - return success anyway
		return nil
	}

	// Invalidate any existing tokens
	s.db.ExecContext(ctx, `
		UPDATE password_reset_tokens 
		SET used_at = $1 
		WHERE user_id = $2 AND used_at IS NULL
	`, time.Now(), userID)

	// Generate secure token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(s.tokenExpiry)

	// Store token
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (id, user_id, token, expires_at, ip, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), userID, token, expiresAt, ip, time.Now())

	if err != nil {
		return fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send email with reset link
	if s.emailService != nil {
		resetLink := fmt.Sprintf("https://app.lco.org.ly/reset-password?token=%s", token)

		err = s.emailService.SendEmail(ctx, email, "Password Reset Request", fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif; direction: rtl;">
				<h2>طلب إعادة تعيين كلمة المرور</h2>
				<p>مرحباً %s،</p>
				<p>لقد تلقينا طلباً لإعادة تعيين كلمة المرور الخاصة بحسابك.</p>
				<p>انقر على الرابط التالي لإعادة تعيين كلمة المرور:</p>
				<a href="%s" style="background-color: #4A90D9; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px;">
					إعادة تعيين كلمة المرور
				</a>
				<p style="margin-top: 20px;">هذا الرابط صالح لمدة 30 دقيقة فقط.</p>
				<p>إذا لم تطلب إعادة تعيين كلمة المرور، يرجى تجاهل هذه الرسالة.</p>
				<hr>
				<p style="color: #888; font-size: 12px;">CSO-ERP System</p>
			</body>
			</html>
		`, fullName, resetLink), true)

		if err != nil {
			// Log error but don't fail - user still has token
			fmt.Printf("Failed to send reset email: %v\n", err)
		}
	}

	return nil
}

// ValidateResetToken validates a password reset token
func (s *PasswordResetService) ValidateResetToken(ctx context.Context, token string) (uuid.UUID, error) {
	var rt PasswordResetToken
	var usedAt *time.Time

	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, used_at
		FROM password_reset_tokens
		WHERE token = $1
	`, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &usedAt)

	if err != nil {
		return uuid.Nil, apperrors.ErrInvalidToken
	}

	if usedAt != nil {
		return uuid.Nil, apperrors.New("RESET_001", "Reset token already used", 400)
	}

	if time.Now().After(rt.ExpiresAt) {
		return uuid.Nil, apperrors.Err2FACodeExpired
	}

	return rt.UserID, nil
}

// ResetPassword resets user password with token
func (s *PasswordResetService) ResetPassword(ctx context.Context, token, newPassword, ip string) error {
	// Validate token
	userID, err := s.ValidateResetToken(ctx, token)
	if err != nil {
		return err
	}

	// Validate password strength
	if len(newPassword) < 8 {
		return apperrors.ErrWeakPassword
	}

	// Hash new password
	hashedPassword, err := security.HashPassword(newPassword, nil)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Start transaction
	tx, err := s.db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update password and increment token version
	_, err = tx.ExecContext(ctx, `
		UPDATE users 
		SET password_hash = $1, 
			token_version = token_version + 1,
			updated_at = $2
		WHERE id = $3
	`, hashedPassword, time.Now(), userID)

	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Mark token as used
	_, err = tx.ExecContext(ctx, `
		UPDATE password_reset_tokens 
		SET used_at = $1 
		WHERE token = $2
	`, time.Now(), token)

	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// CleanupExpiredTokens removes expired tokens
func (s *PasswordResetService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM password_reset_tokens 
		WHERE expires_at < $1 OR used_at IS NOT NULL
	`, time.Now().Add(-24*time.Hour))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
