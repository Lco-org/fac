package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

// TestDB provides a test database connection
type TestDB struct {
	DB *sql.DB
}

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *TestDB {
	t.Helper()

	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/cso-erp_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping test, database not available: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		t.Skipf("Skipping test, database not reachable: %v", err)
		return nil
	}

	return &TestDB{DB: db}
}

// Close closes the test database connection
func (tdb *TestDB) Close() {
	if tdb.DB != nil {
		tdb.DB.Close()
	}
}

// SetupTestUser creates a test user and returns the user ID
func (tdb *TestDB) SetupTestUser(t *testing.T, username, email, passwordHash string) string {
	t.Helper()

	var userID string
	err := tdb.DB.QueryRow(`
		INSERT INTO users (username, email, password_hash, full_name, role, two_factor_enabled, is_active)
		VALUES ($1, $2, $3, 'Test User', 'admin', false, true)
		ON CONFLICT (email) DO UPDATE SET password_hash = $3
		RETURNING id
	`, username, email, passwordHash).Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to setup test user: %v", err)
	}

	return userID
}

// CleanupTestUser removes a test user
func (tdb *TestDB) CleanupTestUser(t *testing.T, email string) {
	t.Helper()
	_, _ = tdb.DB.Exec("DELETE FROM users WHERE email = $1", email)
}

// CleanupAllTestData removes all test data
func (tdb *TestDB) CleanupAllTestData(t *testing.T) {
	t.Helper()

	// Clean in order due to foreign keys
	queries := []string{
		"DELETE FROM otp_codes WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@test.lco.org.ly')",
		"DELETE FROM users WHERE email LIKE '%@test.lco.org.ly'",
	}

	for _, q := range queries {
		_, _ = tdb.DB.Exec(q)
	}
}

// Transaction helper for rollback after tests
type TestTransaction struct {
	Tx *sql.Tx
}

// BeginTx starts a test transaction
func (tdb *TestDB) BeginTx(t *testing.T) *TestTransaction {
	t.Helper()

	tx, err := tdb.DB.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	return &TestTransaction{Tx: tx}
}

// Rollback rolls back the test transaction
func (ttx *TestTransaction) Rollback() {
	if ttx.Tx != nil {
		ttx.Tx.Rollback()
	}
}

// MockContext creates a context with test values
func MockContext() context.Context {
	return context.Background()
}

// AssertNil asserts that an error is nil
func AssertNil(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Errorf("%s: expected nil error, got: %v", msg, err)
	}
}

// AssertNotNil asserts that a value is not nil
func AssertNotNil(t *testing.T, val interface{}, msg string) {
	t.Helper()
	if val == nil {
		t.Errorf("%s: expected non-nil value", msg)
	}
}

// AssertEqual asserts that two values are equal
func AssertEqual(t *testing.T, expected, actual interface{}, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %v, got %v", msg, expected, actual)
	}
}

// AssertTrue asserts that a condition is true
func AssertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true", msg)
	}
}

// AssertFalse asserts that a condition is false
func AssertFalse(t *testing.T, condition bool, msg string) {
	t.Helper()
	if condition {
		t.Errorf("%s: expected false", msg)
	}
}

// AssertError asserts that an error matches expected error
func AssertError(t *testing.T, expected, actual error, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected error %v, got %v", msg, expected, actual)
	}
}

// GenerateTestEmail generates a unique test email
func GenerateTestEmail(prefix string) string {
	return fmt.Sprintf("%s_%d@test.lco.org.ly", prefix, os.Getpid())
}
