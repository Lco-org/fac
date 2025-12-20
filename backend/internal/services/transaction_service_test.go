package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"testing"

	"github.com/google/uuid"
)

// ===============================================
// TRANSACTION SERVICE TESTS
// ===============================================

// setupTestTransactionService creates a TransactionService for testing
func setupTestTransactionService(t *testing.T) (*TransactionService, *database.DB, func()) {
	t.Helper()

	db := setupTestDB(t)
	if db == nil {
		return nil, nil, func() {}
	}

	transactionService := NewTransactionService(db)

	cleanup := func() {
		// Cleanup test transactions
		db.Exec("DELETE FROM transactions WHERE description LIKE '%TEST%'")
		db.Close()
	}

	return transactionService, db, cleanup
}

// setupTestData creates test containers and projects for transaction tests
func setupTestData(t *testing.T, db *database.DB) (containerID, projectID uuid.UUID) {
	t.Helper()

	// Create test container
	err := db.QueryRow(`
		INSERT INTO financial_containers (name, type, currency, balance, is_active)
		VALUES ('TEST Container', 'cash', 'LYD', 10000.00, true)
		ON CONFLICT DO NOTHING
		RETURNING id
	`).Scan(&containerID)

	if err != nil {
		// Container might already exist, try to get it
		db.QueryRow(`SELECT id FROM financial_containers WHERE name = 'TEST Container' LIMIT 1`).Scan(&containerID)
	}

	// Create test project
	err = db.QueryRow(`
		INSERT INTO projects (code, name, currency, is_active)
		VALUES ('TEST001', 'TEST Project', 'LYD', true)
		ON CONFLICT (code) DO UPDATE SET is_active = true
		RETURNING id
	`).Scan(&projectID)

	if err != nil {
		// Project might already exist, try to get it
		db.QueryRow(`SELECT id FROM projects WHERE code = 'TEST001' LIMIT 1`).Scan(&projectID)
	}

	return containerID, projectID
}

// cleanupTestData removes test containers and projects
func cleanupTestData(t *testing.T, db *database.DB) {
	t.Helper()
	db.Exec("DELETE FROM transactions WHERE description LIKE '%TEST%'")
	db.Exec("DELETE FROM projects WHERE code = 'TEST001'")
	db.Exec("DELETE FROM financial_containers WHERE name = 'TEST Container'")
}

// ===============================================
// CREATE TRANSACTION TESTS
// ===============================================

func TestTransactionService_CreateTransaction_Income(t *testing.T) {
	transactionService, db, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(t, db)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	userID := uuid.New()

	req := &models.CreateTransactionRequest{
		Type:         models.TransactionIncomeGrant,
		Amount:       1000.00,
		Currency:     models.CurrencyLYD,
		ExchangeRate: 1.0,
		ContainerID:  containerID,
		ProjectID:    projectID,
		Description:  "TEST Income Grant Transaction",
	}

	resp, err := transactionService.CreateTransaction(ctx, req, userID)

	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response should not be nil")
	}

	if resp.Transaction.Amount != 1000.00 {
		t.Errorf("Expected amount 1000.00, got %f", resp.Transaction.Amount)
	}

	if resp.Transaction.Type != models.TransactionIncomeGrant {
		t.Errorf("Expected type income_grant, got %s", resp.Transaction.Type)
	}

	if resp.Transaction.Status != models.TransactionCompleted {
		t.Errorf("Expected status completed, got %s", resp.Transaction.Status)
	}
}

func TestTransactionService_CreateTransaction_Disbursement(t *testing.T) {
	transactionService, db, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(t, db)
	defer cleanupTestData(t, db)

	// Ensure container has balance
	db.Exec("UPDATE financial_containers SET balance = 10000.00 WHERE id = $1", containerID)

	ctx := context.Background()
	userID := uuid.New()

	req := &models.CreateTransactionRequest{
		Type:         models.TransactionDisbursementOperational,
		Amount:       500.00,
		Currency:     models.CurrencyLYD,
		ExchangeRate: 1.0,
		ContainerID:  containerID,
		ProjectID:    projectID,
		Description:  "TEST Operational Expense",
	}

	resp, err := transactionService.CreateTransaction(ctx, req, userID)

	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	if resp.Transaction.Amount != 500.00 {
		t.Errorf("Expected amount 500.00, got %f", resp.Transaction.Amount)
	}
}

func TestTransactionService_CreateTransaction_InvalidAmount(t *testing.T) {
	transactionService, db, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(t, db)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	userID := uuid.New()

	req := &models.CreateTransactionRequest{
		Type:         models.TransactionIncomeGrant,
		Amount:       0, // Invalid - zero amount
		Currency:     models.CurrencyLYD,
		ExchangeRate: 1.0,
		ContainerID:  containerID,
		ProjectID:    projectID,
		Description:  "TEST Invalid Amount",
	}

	_, err := transactionService.CreateTransaction(ctx, req, userID)

	if err == nil {
		t.Error("Expected error for zero amount, got nil")
	}
}

func TestTransactionService_CreateTransaction_NegativeAmount(t *testing.T) {
	transactionService, db, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(t, db)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	userID := uuid.New()

	req := &models.CreateTransactionRequest{
		Type:         models.TransactionIncomeGrant,
		Amount:       -100.00, // Invalid - negative amount
		Currency:     models.CurrencyLYD,
		ExchangeRate: 1.0,
		ContainerID:  containerID,
		ProjectID:    projectID,
		Description:  "TEST Negative Amount",
	}

	_, err := transactionService.CreateTransaction(ctx, req, userID)

	if err == nil {
		t.Error("Expected error for negative amount, got nil")
	}
}

// ===============================================
// GET TRANSACTION TESTS
// ===============================================

func TestTransactionService_GetTransactionByID(t *testing.T) {
	transactionService, db, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(t, db)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	userID := uuid.New()

	// Create a transaction first
	req := &models.CreateTransactionRequest{
		Type:         models.TransactionIncomeGrant,
		Amount:       750.00,
		Currency:     models.CurrencyLYD,
		ExchangeRate: 1.0,
		ContainerID:  containerID,
		ProjectID:    projectID,
		Description:  "TEST Transaction for GetByID",
	}

	resp, err := transactionService.CreateTransaction(ctx, req, userID)
	if err != nil {
		t.Fatalf("Failed to create test transaction: %v", err)
	}

	// Now get it by ID
	transaction, err := transactionService.GetTransactionByID(ctx, resp.Transaction.ID)

	if err != nil {
		t.Fatalf("GetTransactionByID failed: %v", err)
	}

	if transaction == nil {
		t.Fatal("Transaction should not be nil")
	}

	if transaction.ID != resp.Transaction.ID {
		t.Errorf("Expected ID %s, got %s", resp.Transaction.ID, transaction.ID)
	}

	if transaction.Amount != 750.00 {
		t.Errorf("Expected amount 750.00, got %f", transaction.Amount)
	}
}

func TestTransactionService_GetTransactionByID_NotFound(t *testing.T) {
	transactionService, _, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	randomID := uuid.New()

	_, err := transactionService.GetTransactionByID(ctx, randomID)

	if err == nil {
		t.Error("Expected error for non-existent transaction, got nil")
	}
}

// ===============================================
// REVERSE TRANSACTION TESTS
// ===============================================

func TestTransactionService_ReverseTransaction(t *testing.T) {
	transactionService, db, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(t, db)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	userID := uuid.New()

	// Create a transaction first
	req := &models.CreateTransactionRequest{
		Type:         models.TransactionIncomeGrant,
		Amount:       500.00,
		Currency:     models.CurrencyLYD,
		ExchangeRate: 1.0,
		ContainerID:  containerID,
		ProjectID:    projectID,
		Description:  "TEST Transaction to Reverse",
	}

	resp, err := transactionService.CreateTransaction(ctx, req, userID)
	if err != nil {
		t.Fatalf("Failed to create test transaction: %v", err)
	}

	// Now reverse it
	err = transactionService.ReverseTransaction(ctx, resp.Transaction.ID, "TEST reversal reason", userID)

	if err != nil {
		t.Fatalf("ReverseTransaction failed: %v", err)
	}

	// Verify the transaction is reversed
	transaction, err := transactionService.GetTransactionByID(ctx, resp.Transaction.ID)
	if err != nil {
		t.Fatalf("Failed to get reversed transaction: %v", err)
	}

	if transaction.Status != models.TransactionReversed {
		t.Errorf("Expected status reversed, got %s", transaction.Status)
	}
}

func TestTransactionService_ReverseTransaction_AlreadyReversed(t *testing.T) {
	transactionService, db, cleanup := setupTestTransactionService(t)
	if transactionService == nil {
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(t, db)
	defer cleanupTestData(t, db)

	ctx := context.Background()
	userID := uuid.New()

	// Create and reverse a transaction
	req := &models.CreateTransactionRequest{
		Type:         models.TransactionIncomeGrant,
		Amount:       300.00,
		Currency:     models.CurrencyLYD,
		ExchangeRate: 1.0,
		ContainerID:  containerID,
		ProjectID:    projectID,
		Description:  "TEST Transaction Already Reversed",
	}

	resp, _ := transactionService.CreateTransaction(ctx, req, userID)
	transactionService.ReverseTransaction(ctx, resp.Transaction.ID, "First reversal", userID)

	// Try to reverse again
	err := transactionService.ReverseTransaction(ctx, resp.Transaction.ID, "Second reversal attempt", userID)

	if err == nil {
		t.Error("Expected error when reversing already-reversed transaction, got nil")
	}
}

// ===============================================
// BENCHMARK TESTS
// ===============================================

func BenchmarkCreateTransaction(b *testing.B) {
	transactionService, db, cleanup := setupTestTransactionService(&testing.T{})
	if transactionService == nil {
		b.Skip("Database not available")
		return
	}
	defer cleanup()

	containerID, projectID := setupTestData(&testing.T{}, db)
	defer cleanupTestData(&testing.T{}, db)

	ctx := context.Background()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := &models.CreateTransactionRequest{
			Type:         models.TransactionIncomeGrant,
			Amount:       100.00,
			Currency:     models.CurrencyLYD,
			ExchangeRate: 1.0,
			ContainerID:  containerID,
			ProjectID:    projectID,
			Description:  "TEST Benchmark Transaction",
		}
		transactionService.CreateTransaction(ctx, req, userID)
	}
}
