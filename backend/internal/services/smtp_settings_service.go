package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/google/uuid"
)

// SMTPSettingsService handles SMTP configuration CRUD operations
type SMTPSettingsService struct {
	db  *database.DB
	cfg *config.Config
}

// NewSMTPSettingsService creates a new SMTP settings service
func NewSMTPSettingsService(db *database.DB, cfg *config.Config) *SMTPSettingsService {
	return &SMTPSettingsService{db: db, cfg: cfg}
}

// encryptPassword encrypts SMTP password using AES-256-GCM
func (s *SMTPSettingsService) encryptPassword(password string) (string, error) {
	key := []byte(s.cfg.Security.EncryptionKey)
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(password), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// decryptPassword decrypts SMTP password
func (s *SMTPSettingsService) decryptPassword(encrypted string) (string, error) {
	key := []byte(s.cfg.Security.EncryptionKey)
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// ListSMTPSettings returns all SMTP configurations
func (s *SMTPSettingsService) ListSMTPSettings(ctx context.Context) ([]models.SMTPSettings, error) {
	query := `
		SELECT id, name, smtp_host, smtp_port, username, from_address, from_name,
		       use_tls, is_active, is_default, test_email_sent_at, last_error,
		       created_at, updated_at, created_by
		FROM smtp_settings
		ORDER BY is_default DESC, name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list SMTP settings: %w", err)
	}
	defer rows.Close()

	var settings []models.SMTPSettings
	for rows.Next() {
		var smtp models.SMTPSettings
		err := rows.Scan(
			&smtp.ID, &smtp.Name, &smtp.SMTPHost, &smtp.SMTPPort, &smtp.Username,
			&smtp.FromAddress, &smtp.FromName, &smtp.UseTLS, &smtp.IsActive,
			&smtp.IsDefault, &smtp.TestEmailSentAt, &smtp.LastError,
			&smtp.CreatedAt, &smtp.UpdatedAt, &smtp.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan SMTP settings: %w", err)
		}
		settings = append(settings, smtp)
	}

	return settings, nil
}

// GetDefaultSMTPSettings returns the default active SMTP configuration
func (s *SMTPSettingsService) GetDefaultSMTPSettings(ctx context.Context) (*models.SMTPSettings, error) {
	query := `
		SELECT id, name, smtp_host, smtp_port, username, password_encrypted, 
		       from_address, from_name, use_tls, is_active, is_default, 
		       test_email_sent_at, last_error, created_at, updated_at, created_by
		FROM smtp_settings
		WHERE is_active = true AND is_default = true
		LIMIT 1
	`

	var smtp models.SMTPSettings
	err := s.db.QueryRowContext(ctx, query).Scan(
		&smtp.ID, &smtp.Name, &smtp.SMTPHost, &smtp.SMTPPort, &smtp.Username,
		&smtp.PasswordEncrypted, &smtp.FromAddress, &smtp.FromName, &smtp.UseTLS,
		&smtp.IsActive, &smtp.IsDefault, &smtp.TestEmailSentAt, &smtp.LastError,
		&smtp.CreatedAt, &smtp.UpdatedAt, &smtp.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get default SMTP: %w", err)
	}

	return &smtp, nil
}

// GetSMTPSettingsByID returns SMTP settings by ID
func (s *SMTPSettingsService) GetSMTPSettingsByID(ctx context.Context, id uuid.UUID) (*models.SMTPSettings, error) {
	query := `
		SELECT id, name, smtp_host, smtp_port, username, password_encrypted, 
		       from_address, from_name, use_tls, is_active, is_default, 
		       test_email_sent_at, last_error, created_at, updated_at, created_by
		FROM smtp_settings
		WHERE id = $1
	`

	var smtp models.SMTPSettings
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&smtp.ID, &smtp.Name, &smtp.SMTPHost, &smtp.SMTPPort, &smtp.Username,
		&smtp.PasswordEncrypted, &smtp.FromAddress, &smtp.FromName, &smtp.UseTLS,
		&smtp.IsActive, &smtp.IsDefault, &smtp.TestEmailSentAt, &smtp.LastError,
		&smtp.CreatedAt, &smtp.UpdatedAt, &smtp.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SMTP settings: %w", err)
	}

	return &smtp, nil
}

// CreateSMTPSettings creates new SMTP configuration
func (s *SMTPSettingsService) CreateSMTPSettings(ctx context.Context, req *models.CreateSMTPSettingsRequest, createdBy uuid.UUID) (*models.SMTPSettings, error) {
	// Encrypt password
	var encryptedPassword *string
	if req.Password != "" {
		encrypted, err := s.encryptPassword(req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		encryptedPassword = &encrypted
	}

	// If this is set as default, unset other defaults
	if req.IsDefault {
		_, err := s.db.ExecContext(ctx, "UPDATE smtp_settings SET is_default = false WHERE is_default = true")
		if err != nil {
			return nil, fmt.Errorf("failed to unset default: %w", err)
		}
	}

	query := `
		INSERT INTO smtp_settings (
			name, smtp_host, smtp_port, username, password_encrypted,
			from_address, from_name, use_tls, is_active, is_default, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10)
		RETURNING id, created_at, updated_at
	`

	var smtp models.SMTPSettings
	smtp.Name = req.Name
	smtp.SMTPHost = req.SMTPHost
	smtp.SMTPPort = req.SMTPPort
	smtp.Username = req.Username
	smtp.FromAddress = req.FromAddress
	smtp.FromName = req.FromName
	smtp.UseTLS = req.UseTLS
	smtp.IsActive = true
	smtp.IsDefault = req.IsDefault
	smtp.CreatedBy = &createdBy

	err := s.db.QueryRowContext(ctx, query,
		req.Name, req.SMTPHost, req.SMTPPort, req.Username, encryptedPassword,
		req.FromAddress, req.FromName, req.UseTLS, req.IsDefault, createdBy,
	).Scan(&smtp.ID, &smtp.CreatedAt, &smtp.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP settings: %w", err)
	}

	return &smtp, nil
}

// UpdateSMTPSettings updates SMTP configuration
func (s *SMTPSettingsService) UpdateSMTPSettings(ctx context.Context, id uuid.UUID, req *models.UpdateSMTPSettingsRequest) (*models.SMTPSettings, error) {
	existing, err := s.GetSMTPSettingsByID(ctx, id)
	if err != nil || existing == nil {
		return nil, fmt.Errorf("SMTP settings not found")
	}

	// If setting as default, unset others
	if req.IsDefault != nil && *req.IsDefault {
		_, err := s.db.ExecContext(ctx, "UPDATE smtp_settings SET is_default = false WHERE id != $1", id)
		if err != nil {
			return nil, fmt.Errorf("failed to unset default: %w", err)
		}
	}

	// Build and execute update
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.SMTPHost != nil {
		existing.SMTPHost = *req.SMTPHost
	}
	if req.SMTPPort != nil {
		existing.SMTPPort = *req.SMTPPort
	}
	if req.Username != nil {
		existing.Username = *req.Username
	}
	if req.FromAddress != nil {
		existing.FromAddress = *req.FromAddress
	}
	if req.FromName != nil {
		existing.FromName = req.FromName
	}
	if req.UseTLS != nil {
		existing.UseTLS = *req.UseTLS
	}
	if req.IsDefault != nil {
		existing.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	// Handle password update
	var encryptedPassword *string
	if req.Password != nil && *req.Password != "" {
		encrypted, err := s.encryptPassword(*req.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt password: %w", err)
		}
		encryptedPassword = &encrypted
	}

	updateQuery := `
		UPDATE smtp_settings SET
			name = $1, smtp_host = $2, smtp_port = $3, username = $4,
			from_address = $5, from_name = $6, use_tls = $7, is_default = $8,
			is_active = $9, updated_at = NOW()
		WHERE id = $10
	`

	_, err = s.db.ExecContext(ctx, updateQuery,
		existing.Name, existing.SMTPHost, existing.SMTPPort, existing.Username,
		existing.FromAddress, existing.FromName, existing.UseTLS, existing.IsDefault,
		existing.IsActive, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update SMTP settings: %w", err)
	}

	// Update password separately if provided
	if encryptedPassword != nil {
		_, err = s.db.ExecContext(ctx, "UPDATE smtp_settings SET password_encrypted = $1 WHERE id = $2", *encryptedPassword, id)
		if err != nil {
			return nil, fmt.Errorf("failed to update password: %w", err)
		}
	}

	return s.GetSMTPSettingsByID(ctx, id)
}

// DeleteSMTPSettings soft-deletes SMTP settings
func (s *SMTPSettingsService) DeleteSMTPSettings(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE smtp_settings SET is_active = false, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	return err
}

// GetDecryptedPassword returns the decrypted password for SMTP auth
func (s *SMTPSettingsService) GetDecryptedPassword(ctx context.Context, id uuid.UUID) (string, error) {
	smtp, err := s.GetSMTPSettingsByID(ctx, id)
	if err != nil || smtp == nil {
		return "", fmt.Errorf("SMTP settings not found")
	}

	if smtp.PasswordEncrypted == "" {
		return "", nil
	}

	return s.decryptPassword(smtp.PasswordEncrypted)
}
