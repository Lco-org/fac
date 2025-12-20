package services

import (
	"context"
	"cso-erp/internal/database"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
)

// ===============================================
// TOKEN SERVICE TESTS
// ===============================================

func TestTokenService_GenerateRefreshToken(t *testing.T) {
	tokenService, db, cleanup := setupTestTokenService(t)
	if tokenService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForTokens(t, db)

	// Generate refresh token
	token, err := tokenService.GenerateRefreshToken(ctx, userID, "Chrome/Windows", "127.0.0.1")

	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Error("Token should not be empty")
	}

	if len(token) < 64 {
		t.Errorf("Token should be at least 64 characters, got %d", len(token))
	}
}

func TestTokenService_ValidateRefreshToken_Success(t *testing.T) {
	tokenService, db, cleanup := setupTestTokenService(t)
	if tokenService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForTokens(t, db)

	// Generate token
	token, _ := tokenService.GenerateRefreshToken(ctx, userID, "Chrome/Windows", "127.0.0.1")

	// Validate token
	validatedUserID, err := tokenService.ValidateRefreshToken(ctx, token)

	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}

	if validatedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, validatedUserID)
	}
}

func TestTokenService_ValidateRefreshToken_Invalid(t *testing.T) {
	tokenService, _, cleanup := setupTestTokenService(t)
	if tokenService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// Try to validate invalid token
	_, err := tokenService.ValidateRefreshToken(ctx, "invalid-token-xyz")

	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestTokenService_RevokeRefreshToken(t *testing.T) {
	tokenService, db, cleanup := setupTestTokenService(t)
	if tokenService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForTokens(t, db)

	// Generate and revoke token
	token, _ := tokenService.GenerateRefreshToken(ctx, userID, "Chrome/Windows", "127.0.0.1")
	err := tokenService.RevokeRefreshToken(ctx, token)

	if err != nil {
		t.Fatalf("RevokeRefreshToken failed: %v", err)
	}

	// Try to validate revoked token
	_, err = tokenService.ValidateRefreshToken(ctx, token)

	if err == nil {
		t.Error("Expected error for revoked token, got nil")
	}
}

func TestTokenService_RevokeAllUserTokens(t *testing.T) {
	tokenService, db, cleanup := setupTestTokenService(t)
	if tokenService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForTokens(t, db)

	// Generate multiple tokens
	token1, _ := tokenService.GenerateRefreshToken(ctx, userID, "Chrome/Windows", "127.0.0.1")
	token2, _ := tokenService.GenerateRefreshToken(ctx, userID, "Firefox/Mac", "192.168.1.1")

	// Revoke all
	err := tokenService.RevokeAllUserTokens(ctx, userID)

	if err != nil {
		t.Fatalf("RevokeAllUserTokens failed: %v", err)
	}

	// Both tokens should be invalid
	_, err1 := tokenService.ValidateRefreshToken(ctx, token1)
	_, err2 := tokenService.ValidateRefreshToken(ctx, token2)

	if err1 == nil || err2 == nil {
		t.Error("Expected errors for revoked tokens")
	}
}

func TestTokenService_RotateRefreshToken(t *testing.T) {
	tokenService, db, cleanup := setupTestTokenService(t)
	if tokenService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForTokens(t, db)

	// Generate initial token
	oldToken, _ := tokenService.GenerateRefreshToken(ctx, userID, "Chrome/Windows", "127.0.0.1")

	// Rotate token
	newToken, returnedUserID, err := tokenService.RotateRefreshToken(ctx, oldToken, "Chrome/Windows", "127.0.0.1")

	if err != nil {
		t.Fatalf("RotateRefreshToken failed: %v", err)
	}

	if newToken == oldToken {
		t.Error("New token should be different from old token")
	}

	if returnedUserID != userID {
		t.Errorf("Expected userID %s, got %s", userID, returnedUserID)
	}

	// Old token should be invalid
	_, err = tokenService.ValidateRefreshToken(ctx, oldToken)
	if err == nil {
		t.Error("Expected error for old token after rotation")
	}

	// New token should be valid
	_, err = tokenService.ValidateRefreshToken(ctx, newToken)
	if err != nil {
		t.Errorf("New token should be valid: %v", err)
	}
}

func TestTokenService_GetActiveTokens(t *testing.T) {
	tokenService, db, cleanup := setupTestTokenService(t)
	if tokenService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	userID := createTestUserForTokens(t, db)

	// Generate multiple tokens
	tokenService.GenerateRefreshToken(ctx, userID, "Chrome/Windows", "127.0.0.1")
	tokenService.GenerateRefreshToken(ctx, userID, "Firefox/Mac", "192.168.1.1")

	// Get active tokens
	tokens, err := tokenService.GetActiveTokens(ctx, userID)

	if err != nil {
		t.Fatalf("GetActiveTokens failed: %v", err)
	}

	if len(tokens) != 2 {
		t.Errorf("Expected 2 active tokens, got %d", len(tokens))
	}
}

// ===============================================
// HELPER FUNCTIONS
// ===============================================

func setupTestTokenService(t *testing.T) (*TokenService, *testDB, func()) {
	t.Helper()

	rawDB := setupTestDB(t)
	if rawDB == nil {
		return nil, nil, func() {}
	}

	// Wrap in testDB struct
	db := &testDB{DB: rawDB}

	// Create refresh_tokens table if not exists
	db.Exec(`
		CREATE TABLE IF NOT EXISTS refresh_tokens (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			token VARCHAR(255) NOT NULL UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			is_revoked BOOLEAN DEFAULT FALSE,
			device_info VARCHAR(255),
			ip VARCHAR(45),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		)
	`)

	tokenService := NewTokenService(db.DB, 15*time.Minute, 7*24*time.Hour)

	cleanup := func() {
		db.Exec("DELETE FROM refresh_tokens WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%test%')")
		db.Exec("DELETE FROM users WHERE email LIKE '%test%'")
		db.Close()
	}

	return tokenService, db, cleanup
}

type testDB struct {
	DB *database.DB
}

func (t *testDB) Exec(query string, args ...interface{}) {
	t.DB.Exec(query, args...)
}

func (t *testDB) Close() {
	t.DB.Close()
}

func createTestUserForTokens(t *testing.T, db *testDB) uuid.UUID {
	t.Helper()

	userID := uuid.New()
	email := fmt.Sprintf("tokentest_%s@test.lco.org.ly", uuid.New().String()[:8])

	_, err := db.DB.Exec(`
		INSERT INTO users (id, username, email, password_hash, full_name, role, is_active,token_version)
		VALUES ($1, $2, $3, 'hashedpw', 'Test User', 'admin', true, 1)
		ON CONFLICT DO NOTHING
	`, userID, email, email)

	if err != nil {
		t.Logf("Error creating test user: %v", err)
	}

	return userID
}
