package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"cso-erp/internal/database"
	apperrors "cso-erp/internal/errors"

	"github.com/google/uuid"
)

// TokenService handles access and refresh token operations
type TokenService struct {
	db                 *database.DB
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewTokenService creates a new token service
func NewTokenService(db *database.DB, accessExpiry, refreshExpiry time.Duration) *TokenService {
	return &TokenService{
		db:                 db,
		accessTokenExpiry:  accessExpiry,
		refreshTokenExpiry: refreshExpiry,
	}
}

// RefreshToken represents a refresh token stored in the database
type RefreshToken struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Token      string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	IsRevoked  bool
	DeviceInfo string
	IP         string
}

// GenerateRefreshToken generates a new refresh token for a user
func (s *TokenService) GenerateRefreshToken(ctx context.Context, userID uuid.UUID, deviceInfo, ip string) (string, error) {
	// Generate secure random token
	tokenBytes := make([]byte, 64)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	token := base64.URLEncoding.EncodeToString(tokenBytes)

	expiresAt := time.Now().Add(s.refreshTokenExpiry)

	// Store in database
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, device_info, ip, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New(), userID, token, expiresAt, deviceInfo, ip, time.Now())

	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return token, nil
}

// ValidateRefreshToken validates a refresh token and returns user ID
func (s *TokenService) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	var rt RefreshToken
	err := s.db.QueryRowContext(ctx, `
		SELECT id, user_id, token, expires_at, is_revoked
		FROM refresh_tokens
		WHERE token = $1
	`, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.IsRevoked)

	if err != nil {
		return uuid.Nil, apperrors.ErrInvalidToken
	}

	if rt.IsRevoked {
		return uuid.Nil, apperrors.ErrSessionInvalidated
	}

	if time.Now().After(rt.ExpiresAt) {
		return uuid.Nil, apperrors.ErrInvalidToken
	}

	return rt.UserID, nil
}

// RevokeRefreshToken revokes a specific refresh token
func (s *TokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET is_revoked = true WHERE token = $1
	`, token)
	return err
}

// RevokeAllUserTokens revokes all refresh tokens for a user
func (s *TokenService) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE refresh_tokens SET is_revoked = true WHERE user_id = $1
	`, userID)
	return err
}

// RotateRefreshToken rotates a refresh token (revoke old, create new)
func (s *TokenService) RotateRefreshToken(ctx context.Context, oldToken string, deviceInfo, ip string) (string, uuid.UUID, error) {
	// Validate old token
	userID, err := s.ValidateRefreshToken(ctx, oldToken)
	if err != nil {
		return "", uuid.Nil, err
	}

	// Revoke old token
	if err := s.RevokeRefreshToken(ctx, oldToken); err != nil {
		return "", uuid.Nil, err
	}

	// Generate new token
	newToken, err := s.GenerateRefreshToken(ctx, userID, deviceInfo, ip)
	if err != nil {
		return "", uuid.Nil, err
	}

	return newToken, userID, nil
}

// CleanupExpiredTokens removes expired tokens from database
func (s *TokenService) CleanupExpiredTokens(ctx context.Context) (int64, error) {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM refresh_tokens 
		WHERE expires_at < $1 OR is_revoked = true
	`, time.Now())
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// GetActiveTokens returns all active tokens for a user
func (s *TokenService) GetActiveTokens(ctx context.Context, userID uuid.UUID) ([]RefreshToken, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, token, expires_at, device_info, ip, created_at
		FROM refresh_tokens
		WHERE user_id = $1 AND is_revoked = false AND expires_at > $2
		ORDER BY created_at DESC
	`, userID, time.Now())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []RefreshToken
	for rows.Next() {
		var t RefreshToken
		if err := rows.Scan(&t.ID, &t.UserID, &t.Token, &t.ExpiresAt, &t.DeviceInfo, &t.IP, &t.CreatedAt); err != nil {
			return nil, err
		}
		tokens = append(tokens, t)
	}
	return tokens, nil
}
