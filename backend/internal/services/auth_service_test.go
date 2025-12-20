package services

import (
	"context"
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"cso-erp/internal/security"
	"os"
	"testing"
	"time"
)

// Test configuration
var testDBConfig = &config.DatabaseConfig{
	Host:     getEnv("TEST_DB_HOST", "localhost"),
	Port:     getEnv("TEST_DB_PORT", "5432"),
	User:     getEnv("TEST_DB_USER", "postgres"),
	Password: getEnv("TEST_DB_PASS", "postgres"),
	DBName:   getEnv("TEST_DB_NAME", "cso-erp_test"),
	SSLMode:  "disable",
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// setupTestDB creates a database connection for tests
func setupTestDB(t *testing.T) *database.DB {
	t.Helper()

	db, err := database.Connect(testDBConfig)
	if err != nil {
		t.Skipf("Skipping test, database not available: %v", err)
		return nil
	}

	return db
}

// setupTestAuthService creates an AuthService for testing
func setupTestAuthService(t *testing.T) (*AuthService, *database.DB, func()) {
	t.Helper()

	db := setupTestDB(t)
	if db == nil {
		return nil, nil, func() {}
	}

	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:        "test-secret-key-at-least-32-chars!!",
			Expiry:        time.Hour,
			RefreshExpiry: time.Hour * 24,
		},
	}

	emailService := NewEmailService(nil) // nil config = no-op email
	authService := NewAuthService(db, cfg, emailService)

	cleanup := func() {
		// Cleanup test data
		db.Exec("DELETE FROM otp_codes WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@test.lco.org.ly')")
		db.Exec("DELETE FROM users WHERE email LIKE '%@test.lco.org.ly'")
		db.Close()
	}

	return authService, db, cleanup
}

// createTestUser creates a user for testing
func createTestUser(t *testing.T, db *database.DB, email, password string, twoFactor bool) string {
	t.Helper()

	passwordHash, err := security.HashPassword(password, nil)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	var userID string
	err = db.QueryRow(`
		INSERT INTO users (username, email, password_hash, full_name, role, two_factor_enabled, is_active)
		VALUES ($1, $2, $3, 'Test User', 'admin', $4, true)
		ON CONFLICT (email) DO UPDATE SET password_hash = $3, two_factor_enabled = $4
		RETURNING id
	`, email, email, passwordHash, twoFactor).Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// ===============================================
// LOGIN TESTS
// ===============================================

func TestAuthService_Login_Success(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	// Create test user (2FA disabled for simple login)
	email := "login_success@test.lco.org.ly"
	password := "SecurePassword123!"
	createTestUser(t, db, email, password, false)

	// Test login
	ctx := context.Background()
	result, err := authService.Login(ctx, email, password)

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result == nil {
		t.Fatal("Login result is nil")
	}

	if result.Requires2FA {
		t.Error("Should not require 2FA when disabled")
	}

	if result.User == nil {
		t.Fatal("User should be returned when 2FA is disabled")
	}

	if result.User.Email != email {
		t.Errorf("Expected email %s, got %s", email, result.User.Email)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "login_invalid_pass@test.lco.org.ly"
	createTestUser(t, db, email, "CorrectPassword123!", false)

	ctx := context.Background()
	_, err := authService.Login(ctx, email, "WrongPassword123!")

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	authService, _, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	_, err := authService.Login(ctx, "nonexistent@test.lco.org.ly", "Password123!")

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_Login_InactiveUser(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "inactive_user@test.lco.org.ly"
	password := "SecurePassword123!"

	// Create user and then deactivate
	createTestUser(t, db, email, password, false)
	db.Exec("UPDATE users SET is_active = false WHERE email = $1", email)

	ctx := context.Background()
	_, err := authService.Login(ctx, email, password)

	if err != ErrUserNotActive {
		t.Errorf("Expected ErrUserNotActive, got: %v", err)
	}
}

func TestAuthService_Login_With2FA(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "login_2fa@test.lco.org.ly"
	password := "SecurePassword123!"
	createTestUser(t, db, email, password, true) // 2FA enabled

	ctx := context.Background()
	result, err := authService.Login(ctx, email, password)

	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if !result.Requires2FA {
		t.Error("Should require 2FA when enabled")
	}

	if result.User != nil {
		t.Error("User should be nil when 2FA is required")
	}

	if result.PendingUserID.String() == "00000000-0000-0000-0000-000000000000" {
		t.Error("PendingUserID should be set")
	}
}

// ===============================================
// CREATE USER TESTS
// ===============================================

func TestAuthService_CreateUser_Success(t *testing.T) {
	authService, _, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	email := "create_user@test.lco.org.ly"

	user, err := authService.CreateUser(ctx, "testuser", email, "SecurePassword123!", "Test User", models.RoleAdmin, [16]byte{})

	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	if user == nil {
		t.Fatal("User should not be nil")
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}

	if user.Role != models.RoleAdmin {
		t.Errorf("Expected role admin, got %s", user.Role)
	}

	if !user.TwoFactorEnabled {
		t.Error("2FA should be enabled by default")
	}
}

func TestAuthService_CreateUser_InvalidDomain(t *testing.T) {
	authService, _, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	_, err := authService.CreateUser(ctx, "testuser", "user@gmail.com", "SecurePassword123!", "Test User", models.RoleAdmin, [16]byte{})

	if err != ErrEmailDomainNotAllowed {
		t.Errorf("Expected ErrEmailDomainNotAllowed, got: %v", err)
	}
}

func TestAuthService_CreateUser_WeakPassword(t *testing.T) {
	authService, _, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	_, err := authService.CreateUser(ctx, "testuser", "weak_pass@test.lco.org.ly", "short", "Test User", models.RoleAdmin, [16]byte{})

	if err != ErrWeakPassword {
		t.Errorf("Expected ErrWeakPassword, got: %v", err)
	}
}

// ===============================================
// CHANGE PASSWORD TESTS
// ===============================================

func TestAuthService_ChangePassword_Success(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "change_pass@test.lco.org.ly"
	oldPassword := "OldPassword123!"
	newPassword := "NewPassword456!"
	userID := createTestUser(t, db, email, oldPassword, false)

	ctx := context.Background()

	// Parse UUID
	var uid [16]byte
	copy(uid[:], userID)

	err := authService.ChangePassword(ctx, uid, oldPassword, newPassword)
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	// Verify new password works
	result, err := authService.Login(ctx, email, newPassword)
	if err != nil {
		t.Fatalf("Login with new password failed: %v", err)
	}

	if result.User == nil {
		t.Error("Should be able to login with new password")
	}
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "change_pass_wrong@test.lco.org.ly"
	oldPassword := "OldPassword123!"
	userID := createTestUser(t, db, email, oldPassword, false)

	ctx := context.Background()

	var uid [16]byte
	copy(uid[:], userID)

	err := authService.ChangePassword(ctx, uid, "WrongOldPassword!", "NewPassword456!")

	if err != ErrInvalidCredentials {
		t.Errorf("Expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestAuthService_ChangePassword_WeakNewPassword(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "change_pass_weak@test.lco.org.ly"
	oldPassword := "OldPassword123!"
	userID := createTestUser(t, db, email, oldPassword, false)

	ctx := context.Background()

	var uid [16]byte
	copy(uid[:], userID)

	err := authService.ChangePassword(ctx, uid, oldPassword, "short")

	if err != ErrWeakPassword {
		t.Errorf("Expected ErrWeakPassword, got: %v", err)
	}
}

// ===============================================
// GET USER TESTS
// ===============================================

func TestAuthService_GetUserByID_Success(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "get_user@test.lco.org.ly"
	userID := createTestUser(t, db, email, "Password123!", false)

	ctx := context.Background()

	// Parse UUID properly
	var uid [16]byte
	copy(uid[:], userID)

	user, err := authService.GetUserByID(ctx, uid)
	if err != nil {
		t.Fatalf("GetUserByID failed: %v", err)
	}

	if user.Email != email {
		t.Errorf("Expected email %s, got %s", email, user.Email)
	}
}

// ===============================================
// 2FA TOGGLE TESTS
// ===============================================

func TestAuthService_Toggle2FA(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "toggle_2fa@test.lco.org.ly"
	userID := createTestUser(t, db, email, "Password123!", true) // Start with 2FA enabled

	ctx := context.Background()

	var uid [16]byte
	copy(uid[:], userID)

	// Disable 2FA
	err := authService.Toggle2FA(ctx, uid, false)
	if err != nil {
		t.Fatalf("Toggle2FA (disable) failed: %v", err)
	}

	// Verify login doesn't require 2FA
	result, err := authService.Login(ctx, email, "Password123!")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if result.Requires2FA {
		t.Error("2FA should be disabled")
	}

	// Re-enable 2FA
	err = authService.Toggle2FA(ctx, uid, true)
	if err != nil {
		t.Fatalf("Toggle2FA (enable) failed: %v", err)
	}
}

// ===============================================
// REVOKE SESSIONS TESTS
// ===============================================

func TestAuthService_RevokeAllSessions(t *testing.T) {
	authService, db, cleanup := setupTestAuthService(t)
	if authService == nil {
		return
	}
	defer cleanup()

	email := "revoke_sessions@test.lco.org.ly"
	userID := createTestUser(t, db, email, "Password123!", false)

	ctx := context.Background()

	var uid [16]byte
	copy(uid[:], userID)

	// Get initial token version
	var initialVersion int
	db.QueryRow("SELECT token_version FROM users WHERE id = $1", userID).Scan(&initialVersion)

	// Revoke sessions
	err := authService.RevokeAllSessions(ctx, uid)
	if err != nil {
		t.Fatalf("RevokeAllSessions failed: %v", err)
	}

	// Verify token version increased
	var newVersion int
	db.QueryRow("SELECT token_version FROM users WHERE id = $1", userID).Scan(&newVersion)

	if newVersion <= initialVersion {
		t.Error("Token version should increase after revoking sessions")
	}
}

// ===============================================
// BENCHMARK TESTS
// ===============================================

func BenchmarkLogin(b *testing.B) {
	authService, db, cleanup := setupTestAuthService(&testing.T{})
	if authService == nil {
		b.Skip("Database not available")
		return
	}
	defer cleanup()

	email := "bench_login@test.lco.org.ly"
	password := "SecurePassword123!"
	createTestUser(&testing.T{}, db, email, password, false)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authService.Login(ctx, email, password)
	}
}
