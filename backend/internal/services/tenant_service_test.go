package services

import (
	"context"
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// ===============================================
// TENANT SERVICE TESTS
// ===============================================

// Master DB config for tenant service tests
var testMasterDBConfig = &config.DatabaseConfig{
	Host:     getEnv("TEST_DB_HOST", "localhost"),
	Port:     getEnv("TEST_DB_PORT", "5432"),
	User:     getEnv("TEST_DB_USER", "postgres"),
	Password: getEnv("TEST_DB_PASS", "postgres"),
	DBName:   getEnv("TEST_DB_NAME", "erp_master"),
	SSLMode:  "disable",
}

// setupTestTenantService creates a TenantService for testing
func setupTestTenantService(t *testing.T) (*TenantService, *database.DB, func()) {
	t.Helper()

	// Load .env
	_ = godotenv.Load("../../.env")

	// Helper to get env with fallback
	getEnv := func(key, fallback string) string {
		if value, ok := os.LookupEnv(key); ok {
			return value
		}
		return fallback
	}

	// Update config from env (prefer TEST_DB_*, fallback to DB_*)
	dbConfig := &config.DatabaseConfig{
		Host:     getEnv("TEST_DB_HOST", getEnv("DB_HOST", "localhost")),
		Port:     getEnv("TEST_DB_PORT", getEnv("DB_PORT", "5432")),
		User:     getEnv("TEST_DB_USER", getEnv("DB_USER", "postgres")),
		Password: getEnv("TEST_DB_PASS", getEnv("DB_PASSWORD", "postgres")),
		DBName:   getEnv("TEST_DB_NAME", getEnv("DB_NAME", "erp_master")),
		SSLMode:  "disable",
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		t.Skipf("Skipping test, master database not available: %v", err)
		return nil, nil, func() {}
	}

	cfg := &config.Config{
		MultiTenant: config.MultiTenantConfig{
			MasterDBName:            dbConfig.DBName,
			TemplateDBName:          "cso-erp",
			TenantDBPrefix:          "test_",
			VerificationCodeExpiry:  5 * time.Minute,
			MaxVerificationAttempts: 3,
		},
		Database: *dbConfig,
	}

	tenantService := NewTenantService(db, cfg)

	// Create default plan for tests
	_, err = tenantService.CreateSubscriptionPlan(context.Background(), &models.CreateSubscriptionPlanRequest{
		Name:         "TEST Setup Default",
		Code:         "TEST_SETUP_DEFAULT",
		IsActive:     true,
		PriceMonthly: 1,
		Currency:     "USD",
		MaxUsers:     10,
		MaxStorageGB: 10,
	})
	if err != nil {
		t.Fatalf("Setup warning: failed to create default plan: %v", err)
	}

	cleanup := func() {
		// Cleanup test data
		db.Exec("DELETE FROM domain_verifications WHERE organization_id IN (SELECT id FROM organizations WHERE name LIKE 'TEST%')")
		db.Exec("DELETE FROM allowed_domains WHERE organization_id IN (SELECT id FROM organizations WHERE name LIKE 'TEST%')")
		db.Exec("DELETE FROM organizations WHERE name LIKE 'TEST%'")
		db.Exec("DELETE FROM subscription_plans WHERE code LIKE 'TEST%'")
		db.Close()
	}

	return tenantService, db, cleanup
}

// ===============================================
// HELPER FUNCTION TESTS
// ===============================================

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Simple name", "Acme Corp", "acme-corp"},
		{"Arabic name", "شركة أكمي", ""},
		{"Mixed", "Company أكمي Inc", "company-inc"},
		{"Special chars", "Test!@#$%Company", "testcompany"},
		{"Multiple spaces", "Test   Company   Name", "test-company-name"},
		{"Already slug", "test-company", "test-company"},
		{"Numbers", "Company 123", "company-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateSlug(tt.input)
			if result != tt.expected {
				t.Errorf("generateSlug(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateRandomCode(t *testing.T) {
	// Test length
	code := generateRandomCode(6)
	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}

	// Test uniqueness
	codes := make(map[string]bool)
	for i := 0; i < 100; i++ {
		c := generateRandomCode(6)
		if codes[c] {
			t.Log("Duplicate code generated (unlikely but possible)")
		}
		codes[c] = true
	}

	// Test character set (should not contain confusing characters like 0, O, I, L, 1)
	for _, char := range code {
		if char == '0' || char == 'O' || char == 'I' || char == 'L' || char == '1' {
			t.Errorf("Code contains confusing character: %c", char)
		}
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	password := generateRandomPassword(12)

	if len(password) != 12 {
		t.Errorf("Expected password length 12, got %d", len(password))
	}

	// Test uniqueness
	passwords := make(map[string]bool)
	for i := 0; i < 100; i++ {
		p := generateRandomPassword(12)
		if passwords[p] {
			t.Error("Duplicate password generated")
		}
		passwords[p] = true
	}
}

func TestSanitizeDBName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"valid_db_name", "valid_db_name"},
		{"valid-db-name", "valid-db-name"},
		{"db!@#$name", "dbname"},
		{"db name", "dbname"},
		{"db.name", "dbname"},
		{"db;DROP TABLE--", "dbDROPTABLE--"},
	}

	for _, tt := range tests {
		result := sanitizeDBName(tt.input)
		if result != tt.expected {
			t.Errorf("sanitizeDBName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestHashVerificationCode(t *testing.T) {
	code1 := "ABC123"
	code2 := "ABC123"
	code3 := "XYZ789"

	hash1 := hashVerificationCode(code1)
	hash2 := hashVerificationCode(code2)
	hash3 := hashVerificationCode(code3)

	// Same input should produce same hash
	if hash1 != hash2 {
		t.Error("Same code should produce same hash")
	}

	// Different input should produce different hash
	if hash1 == hash3 {
		t.Error("Different codes should produce different hashes")
	}

	// Hash should be non-empty
	if len(hash1) == 0 {
		t.Error("Hash should not be empty")
	}
}

// ===============================================
// ORGANIZATION MANAGEMENT TESTS
// ===============================================

func TestTenantService_CreateOrganization(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	var count int
	_ = tenantService.masterDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM subscription_plans WHERE code = 'TEST_SETUP_DEFAULT'").Scan(&count)
	t.Logf("DEBUG: Plans with code TEST_SETUP_DEFAULT: %d", count)

	var total int
	_ = tenantService.masterDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM subscription_plans").Scan(&total)
	t.Logf("DEBUG: Total plans: %d", total)

	req := &models.CreateOrganizationRequest{
		Name:         "TEST Organization " + uuid.New().String()[:8],
		Domain:       "test-org-" + uuid.New().String()[:8] + ".com",
		ContactEmail: "admin@test.com",
	}

	org, err := tenantService.CreateOrganization(ctx, req, createdBy)

	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	if org == nil {
		t.Fatal("Organization should not be nil")
	}

	if org.Name != "TEST Organization" {
		t.Errorf("Expected name 'TEST Organization', got %s", org.Name)
	}

	if org.Status != models.OrgStatusPendingVerification {
		t.Errorf("Expected status pending_verification, got %s", org.Status)
	}

	if org.Slug == "" {
		t.Error("Slug should not be empty")
	}
}

func TestTenantService_CreateOrganization_DuplicateDomain(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()
	domain := "test-dup-" + uuid.New().String()[:8] + ".com"

	req := &models.CreateOrganizationRequest{
		Name:         "TEST First Org",
		Domain:       domain,
		ContactEmail: "admin@test.com",
	}

	// Create first organization
	_, err := tenantService.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("First CreateOrganization failed: %v", err)
	}

	// Try to create second with same domain
	req2 := &models.CreateOrganizationRequest{
		Name:         "TEST Second Org",
		Domain:       domain, // Same domain
		ContactEmail: "admin2@test.com",
	}

	_, err = tenantService.CreateOrganization(ctx, req2, createdBy)
	if err == nil {
		t.Error("Expected error for duplicate domain, got nil")
	}
}

func TestTenantService_GetOrganizationByID(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// Create an organization first
	req := &models.CreateOrganizationRequest{
		Name:         "TEST Get Org",
		Domain:       "test-get-" + uuid.New().String()[:8] + ".com",
		ContactEmail: "admin@test.com",
	}

	created, err := tenantService.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Get it by ID
	org, err := tenantService.GetOrganizationByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetOrganizationByID failed: %v", err)
	}

	if org.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, org.ID)
	}

	if org.Name != "TEST Get Org" {
		t.Errorf("Expected name 'TEST Get Org', got %s", org.Name)
	}
}

func TestTenantService_GetOrganizationByID_NotFound(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	randomID := uuid.New()

	_, err := tenantService.GetOrganizationByID(ctx, randomID)
	if err == nil {
		t.Error("Expected error for non-existent org, got nil")
	}
}

func TestTenantService_ListOrganizations(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// Create a few test organizations
	for i := 0; i < 3; i++ {
		req := &models.CreateOrganizationRequest{
			Name:         "TEST List Org " + uuid.New().String()[:4],
			Domain:       "test-list-" + uuid.New().String()[:8] + ".com",
			ContactEmail: "admin@test.com",
		}
		tenantService.CreateOrganization(ctx, req, createdBy)
	}

	// List all organizations
	orgs, total, err := tenantService.ListOrganizations(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("ListOrganizations failed: %v", err)
	}

	if len(orgs) == 0 {
		t.Error("Expected at least some organizations")
	}

	if total < len(orgs) {
		t.Error("Total should be >= returned items")
	}
}

func TestTenantService_ListOrganizations_WithStatusFilter(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// List only pending organizations
	orgs, _, err := tenantService.ListOrganizations(ctx, "pending_verification", 10, 0)
	if err != nil {
		t.Fatalf("ListOrganizations with filter failed: %v", err)
	}

	// All returned orgs should have pending status
	for _, org := range orgs {
		if org.Status != "pending_verification" {
			t.Errorf("Expected status pending_verification, got %s", org.Status)
		}
	}
}

// ===============================================
// VERIFICATION TESTS
// ===============================================

func TestTenantService_GenerateVerificationCode(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// Create an organization
	req := &models.CreateOrganizationRequest{
		Name:         "TEST Verify Org",
		Domain:       "test-verify-" + uuid.New().String()[:8] + ".com",
		ContactEmail: "admin@test.com",
	}

	org, err := tenantService.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("Failed to create org: %v", err)
	}

	// Generate verification code
	code, err := tenantService.GenerateVerificationCode(ctx, org.ID, "admin@test.com", "127.0.0.1")
	if err != nil {
		t.Fatalf("GenerateVerificationCode failed: %v", err)
	}

	if len(code) != 6 {
		t.Errorf("Expected code length 6, got %d", len(code))
	}
}

func TestTenantService_VerifyCode_Success(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// Create an organization
	req := &models.CreateOrganizationRequest{
		Name:         "TEST Verify Success " + uuid.New().String()[:8],
		Domain:       "test-vs-" + uuid.New().String()[:8] + ".com",
		ContactEmail: "admin@test.com",
	}

	org, err := tenantService.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	// Generate code
	code, _ := tenantService.GenerateVerificationCode(ctx, org.ID, "admin@test.com", "127.0.0.1")

	// Verify code
	err = tenantService.VerifyCode(ctx, org.ID, code, "127.0.0.1")
	if err != nil {
		t.Fatalf("VerifyCode failed: %v", err)
	}

	// Check org status changed to active
	updatedOrg, _ := tenantService.GetOrganizationByID(ctx, org.ID)
	if updatedOrg.Status != models.OrgStatusActive {
		t.Errorf("Expected org status active, got %s", updatedOrg.Status)
	}
}

func TestTenantService_VerifyCode_InvalidCode(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// Create and generate code
	req := &models.CreateOrganizationRequest{
		Name:         "TEST Invalid Code " + uuid.New().String()[:8],
		Domain:       "test-ic-" + uuid.New().String()[:8] + ".com",
		ContactEmail: "admin@test.com",
	}

	org, err := tenantService.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}
	tenantService.GenerateVerificationCode(ctx, org.ID, "admin@test.com", "127.0.0.1")

	// Try to verify with wrong code
	err = tenantService.VerifyCode(ctx, org.ID, "WRONG1", "127.0.0.1")
	if err == nil {
		t.Error("Expected error for invalid code, got nil")
	}
}

// ===============================================
// SUSPEND/ACTIVATE TESTS
// ===============================================

func TestTenantService_SuspendOrganization(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// Create and activate an organization
	req := &models.CreateOrganizationRequest{
		Name:         "TEST Suspend Org " + uuid.New().String()[:8],
		Domain:       "test-sus-" + uuid.New().String()[:8] + ".com",
		ContactEmail: "admin@test.com",
	}

	org, err := tenantService.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}
	code, _ := tenantService.GenerateVerificationCode(ctx, org.ID, "admin@test.com", "127.0.0.1")
	tenantService.VerifyCode(ctx, org.ID, code, "127.0.0.1")

	// Suspend
	err = tenantService.SuspendOrganization(ctx, org.ID, "Test suspension")
	if err != nil {
		t.Fatalf("SuspendOrganization failed: %v", err)
	}

	// Verify status
	updated, _ := tenantService.GetOrganizationByID(ctx, org.ID)
	if updated.Status != models.OrgStatusSuspended {
		t.Errorf("Expected status suspended, got %s", updated.Status)
	}
}

func TestTenantService_ActivateOrganization(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// Create, verify, and suspend
	req := &models.CreateOrganizationRequest{
		Name:         "TEST Activate Org " + uuid.New().String()[:8],
		Domain:       "test-act-" + uuid.New().String()[:8] + ".com",
		ContactEmail: "admin@test.com",
	}

	org, err := tenantService.CreateOrganization(ctx, req, createdBy)
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}
	code, _ := tenantService.GenerateVerificationCode(ctx, org.ID, "admin@test.com", "127.0.0.1")
	tenantService.VerifyCode(ctx, org.ID, code, "127.0.0.1")
	tenantService.SuspendOrganization(ctx, org.ID, "Test")

	// Activate
	err = tenantService.ActivateOrganization(ctx, org.ID)
	if err != nil {
		t.Fatalf("ActivateOrganization failed: %v", err)
	}

	// Verify status
	updated, _ := tenantService.GetOrganizationByID(ctx, org.ID)
	if updated.Status != models.OrgStatusActive {
		t.Errorf("Expected status active, got %s", updated.Status)
	}
}

// ===============================================
// STATS TESTS
// ===============================================

func TestTenantService_GetOrganizationStats(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	stats, err := tenantService.GetOrganizationStats(ctx)
	if err != nil {
		t.Fatalf("GetOrganizationStats failed: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	// Just verify the fields are populated (not negative)
	if stats.TotalOrganizations < 0 {
		t.Error("Total should not be negative")
	}

	if stats.ActiveOrganizations < 0 {
		t.Error("Active should not be negative")
	}
}

// ===============================================
// SUBSCRIPTION PLAN TESTS
// ===============================================

func TestTenantService_CreateSubscriptionPlan(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	req := &models.CreateSubscriptionPlanRequest{
		Name:         "TEST Plan",
		Code:         "TEST_PLAN_001",
		PriceMonthly: 100.0,
		PriceYearly:  1000.0,
		Currency:     "USD",
		MaxUsers:     10,
		MaxStorageGB: 50,
		Features:     []string{"feature1", "feature2"},
		IsActive:     true,
	}

	plan, err := tenantService.CreateSubscriptionPlan(ctx, req)
	if err != nil {
		t.Fatalf("CreateSubscriptionPlan failed: %v", err)
	}

	if plan == nil {
		t.Fatal("Plan should not be nil")
	}

	if plan.Code != "TEST_PLAN_001" {
		t.Errorf("Expected code 'TEST_PLAN_001', got %s", plan.Code)
	}

	if len(plan.Features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(plan.Features))
	}
}

func TestTenantService_UpdateSubscriptionPlan(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()

	// Create a plan
	req := &models.CreateSubscriptionPlanRequest{
		Name:         "TEST Plan Update",
		Code:         "TEST_UPDATE_001",
		Features:     []string{"f1"},
		IsActive:     true,
		Currency:     "USD",
		MaxUsers:     5,
		MaxStorageGB: 10,
	}
	plan, _ := tenantService.CreateSubscriptionPlan(ctx, req)

	// Update it
	newName := "TEST Plan Updated"
	newPrice := 200.0
	newStatus := false

	updateReq := &models.UpdateSubscriptionPlanRequest{
		Name:         &newName,
		PriceMonthly: &newPrice,
		IsActive:     &newStatus,
	}

	updated, err := tenantService.UpdateSubscriptionPlan(ctx, plan.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateSubscriptionPlan failed: %v", err)
	}

	if updated.Name != newName {
		t.Errorf("Expected checked name '%s', got '%s'", newName, updated.Name)
	}

	if updated.PriceMonthly != newPrice {
		t.Errorf("Expected price %f, got %f", newPrice, updated.PriceMonthly)
	}

	if updated.IsActive != false {
		t.Error("Expected IsActive to be false")
	}
}

func TestTenantService_UpdateOrganizationSubscription(t *testing.T) {
	tenantService, _, cleanup := setupTestTenantService(t)
	if tenantService == nil {
		return
	}
	defer cleanup()

	ctx := context.Background()
	createdBy := uuid.New()

	// 1. Create a Plan
	planReq := &models.CreateSubscriptionPlanRequest{
		Name:         "TEST Org Plan",
		Code:         "TEST_ORG_PLAN",
		Features:     []string{"f1"},
		IsActive:     true,
		Currency:     "LYD",
		MaxUsers:     100,
		MaxStorageGB: 100,
	}
	_, _ = tenantService.CreateSubscriptionPlan(ctx, planReq)

	// 2. Create an Organization
	orgReq := &models.CreateOrganizationRequest{
		Name:                 "TEST Sub Org " + uuid.New().String()[:8],
		Domain:               "test-sub-" + uuid.New().String()[:8] + ".com",
		ContactEmail:         "admin@test.com",
		SubscriptionPlanCode: "TEST_ORG_PLAN",
	}

	org, err := tenantService.CreateOrganization(ctx, orgReq, createdBy)
	if err != nil {
		t.Fatalf("Failed to create org with plan: %v", err)
	}

	// 3. Create NEW Plan
	planReq2 := &models.CreateSubscriptionPlanRequest{
		Name:         "TEST New Plan",
		Code:         "TEST_NEW_PLAN",
		Currency:     "LYD",
		MaxUsers:     200,
		MaxStorageGB: 200,
	}
	newPlan, _ := tenantService.CreateSubscriptionPlan(ctx, planReq2)

	// 4. Update Organization Subscription to use NEW Plan
	status := "suspended"
	updateReq := &models.UpdateOrganizationSubscriptionRequest{
		SubscriptionPlanID: newPlan.ID.String(),
		SubscriptionStatus: status,
	}

	err = tenantService.UpdateOrganizationSubscription(ctx, org.ID, updateReq)
	if err != nil {
		t.Fatalf("UpdateOrganizationSubscription failed: %v", err)
	}

	// 5. Verify Update
	updatedOrg, _ := tenantService.GetOrganizationByID(ctx, org.ID)
	if updatedOrg.SubscriptionPlanID == nil || *updatedOrg.SubscriptionPlanID != newPlan.ID {
		t.Errorf("Expected plan ID %s, got %v", newPlan.ID, updatedOrg.SubscriptionPlanID)
	}
	if updatedOrg.SubscriptionStatus != "suspended" {
		t.Errorf("Expected status suspended, got %s", updatedOrg.SubscriptionStatus)
	}
}
