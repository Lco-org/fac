package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/models"
	"cso-erp/internal/security"

	"github.com/google/uuid"
)

// TenantService handles multi-tenant operations for Super System Admin
type TenantService struct {
	masterDB *database.DB
	cfg      *config.Config
}

// NewTenantService creates a new tenant service
func NewTenantService(masterDB *database.DB, cfg *config.Config) *TenantService {
	return &TenantService{
		masterDB: masterDB,
		cfg:      cfg,
	}
}

// ============================================================================
// ORGANIZATION MANAGEMENT
// ============================================================================

// CreateOrganization creates a new organization
func (s *TenantService) CreateOrganization(ctx context.Context, req *models.CreateOrganizationRequest, createdBy uuid.UUID) (*models.Organization, error) {
	// Generate slug from name
	slug := generateSlug(req.Name)

	// Generate database name
	dbName := fmt.Sprintf("%s%s_erp", s.cfg.MultiTenant.TenantDBPrefix, slug)

	// Strict Domain Validation Logic
	emailParts := strings.Split(req.ContactEmail, "@")
	if len(emailParts) != 2 {
		return nil, fmt.Errorf("invalid contact email format")
	}
	emailDomain := strings.ToLower(emailParts[1])
	siteDomain := strings.ToLower(req.Domain)

	// Determine matching strategy
	var allowedDomains []string
	if emailDomain == siteDomain {
		// Scenario 1: Exact Match (e.g. admin@ctf.com.ly matches ctf.com.ly)
		allowedDomains = []string{siteDomain}
	} else if strings.HasSuffix(emailDomain, "."+siteDomain) {
		// Scenario 2: Subdomain Match (e.g. admin@mm.ctf.com.ly matches ctf.com.ly)
		// Add BOTH site domain (root) and email domain (proven subdomain)
		allowedDomains = []string{siteDomain, emailDomain}
	} else {
		// Scenario 3: No Match (REJECT)
		return nil, fmt.Errorf("security violation: admin email domain (%s) must belong to the organization domain (%s)", emailDomain, siteDomain)
	}

	// Check if domain already exists
	var exists bool
	err := s.masterDB.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM organizations WHERE primary_domain = $1 OR database_name = $2)",
		req.Domain, dbName,
	).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check organization existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("organization with this domain or database name already exists")
	}

	// Get Subscription Plan
	var planID uuid.UUID
	if req.SubscriptionPlanCode != "" {
		err = s.masterDB.QueryRowContext(ctx, "SELECT id FROM subscription_plans WHERE code = $1", req.SubscriptionPlanCode).Scan(&planID)
	} else {
		// Fallback to default/cheapest active plan
		err = s.masterDB.QueryRowContext(ctx, "SELECT id FROM subscription_plans WHERE is_active = true ORDER BY price_monthly ASC LIMIT 1").Scan(&planID)
	}
	if err != nil {
		if err == sql.ErrNoRows {
			// Fallback to basic if not found or error, but better to error out
			return nil, fmt.Errorf("invalid subscription plan code")
		}
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	now := time.Now()
	startDate := now
	if req.SubscriptionStart != nil {
		startDate = *req.SubscriptionStart
	}

	var expiresAt *time.Time
	if req.SubscriptionEnd != nil {
		expiresAt = req.SubscriptionEnd
	} else {
		// Default 1 year if not provided? Or keep nil for lifetime?
		// Let's rely on frontend or default logic.
		// If basic (free), maybe no expiry. If pro, maybe 1 month?
		// For now, let's just use what's passed or nil.
	}

	org := &models.Organization{
		ID:                    uuid.New(),
		Name:                  req.Name,
		NameAr:                req.NameAr,
		Slug:                  slug,
		PrimaryDomain:         req.Domain,
		DatabaseName:          dbName,
		ContactEmail:          req.ContactEmail,
		ContactPhone:          req.ContactPhone,
		Status:                models.OrgStatusPendingVerification,
		SubscriptionPlanID:    &planID,
		SubscriptionStatus:    "active", // Default to active if created by SSA
		SubscriptionStartDate: &startDate,
		SubscriptionExpires:   expiresAt,
		CreatedAt:             now,
		UpdatedAt:             now,
		CreatedBy:             createdBy,
	}

	// Insert organization
	_, err = s.masterDB.ExecContext(ctx, `
		INSERT INTO organizations (
			id, name, name_ar, slug, primary_domain, database_name,
			contact_email, contact_phone, status,
			subscription_plan_id, subscription_status, subscription_start_date, subscription_expires_at,
			created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, org.ID, org.Name, org.NameAr, org.Slug, org.PrimaryDomain, org.DatabaseName,
		org.ContactEmail, org.ContactPhone, org.Status,
		org.SubscriptionPlanID, org.SubscriptionStatus, org.SubscriptionStartDate, org.SubscriptionExpires,
		org.CreatedAt, org.UpdatedAt, org.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Insert allowed domains
	for _, domain := range allowedDomains {
		isPrimary := (domain == siteDomain)
		// Logic: if site domain is inserted, it's primary.
		// If subdomain is inserted, it's secondary (is_primary=false).
		// Wait, loop:

		_, err = s.masterDB.ExecContext(ctx, `
			INSERT INTO allowed_domains (id, organization_id, domain, is_primary, is_active, verified, verified_at, verification_method)
			VALUES ($1, $2, $3, $4, true, true, $5, 'admin_email_proof')
		`, uuid.New(), org.ID, domain, isPrimary, now)
		if err != nil {
			return nil, fmt.Errorf("failed to add domain %s: %w", domain, err)
		}
	}

	return org, nil
}

// GenerateVerificationCode generates a 6-character verification code
func (s *TenantService) GenerateVerificationCode(ctx context.Context, orgID uuid.UUID, email string, requestIP string) (string, error) {
	// Generate 6-char alphanumeric code
	code := generateRandomCode(6)
	hashedCode := hashVerificationCode(code)

	expiresAt := time.Now().Add(s.cfg.MultiTenant.VerificationCodeExpiry)

	// Delete any existing unverified codes for this org
	_, _ = s.masterDB.ExecContext(ctx,
		"DELETE FROM domain_verifications WHERE organization_id = $1 AND verified = false",
		orgID)

	// Insert new verification record
	_, err := s.masterDB.ExecContext(ctx, `
		INSERT INTO domain_verifications (
			id, organization_id, email, code_hash, expires_at, request_ip
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New(), orgID, email, hashedCode, expiresAt, requestIP)
	if err != nil {
		return "", fmt.Errorf("failed to create verification record: %w", err)
	}

	return code, nil
}

// VerifyCode verifies the email verification code
func (s *TenantService) VerifyCode(ctx context.Context, orgID uuid.UUID, code string, verifyIP string) error {
	hashedCode := hashVerificationCode(code)

	var verification models.DomainVerification
	err := s.masterDB.QueryRowContext(ctx, `
		SELECT id, organization_id, email, code_hash, expires_at, verified, attempts, max_attempts
		FROM domain_verifications
		WHERE organization_id = $1 AND verified = false
		ORDER BY created_at DESC
		LIMIT 1
	`, orgID).Scan(
		&verification.ID, &verification.OrganizationID, &verification.Email,
		&verification.HashCode, &verification.ExpiresAt, &verification.Verified,
		&verification.Attempts, &verification.MaxAttempts,
	)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no pending verification found")
	}
	if err != nil {
		return fmt.Errorf("failed to get verification: %w", err)
	}

	// Check expiry
	if time.Now().After(verification.ExpiresAt) {
		return fmt.Errorf("verification code has expired")
	}

	// Check max attempts
	if verification.Attempts >= verification.MaxAttempts {
		return fmt.Errorf("maximum verification attempts exceeded")
	}

	// Increment attempts
	_, _ = s.masterDB.ExecContext(ctx,
		"UPDATE domain_verifications SET attempts = attempts + 1 WHERE id = $1",
		verification.ID)

	// Compare codes
	if verification.HashCode != hashedCode {
		return fmt.Errorf("invalid verification code")
	}

	// Mark as verified
	now := time.Now()
	_, err = s.masterDB.ExecContext(ctx, `
		UPDATE domain_verifications 
		SET verified = true, verified_at = $1, verify_ip = $2
		WHERE id = $3
	`, now, verifyIP, verification.ID)
	if err != nil {
		return fmt.Errorf("failed to update verification: %w", err)
	}

	// Update organization status
	_, err = s.masterDB.ExecContext(ctx, `
		UPDATE organizations 
		SET status = $1, verified_at = $2, updated_at = $3
		WHERE id = $4
	`, models.OrgStatusActive, now, now, orgID)
	if err != nil {
		return fmt.Errorf("failed to activate organization: %w", err)
	}

	// Update allowed_domains
	_, _ = s.masterDB.ExecContext(ctx, `
		UPDATE allowed_domains 
		SET verified = true, verified_at = $1, verification_method = 'email'
		WHERE organization_id = $2 AND is_primary = true
	`, now, orgID)

	return nil
}

// CreateTenantDatabase creates a new database for the tenant
func (s *TenantService) CreateTenantDatabase(ctx context.Context, dbName string) error {
	// Connect to postgres database to create new database
	masterConn := s.masterDB.DB

	// Check if database exists
	var exists bool
	err := masterConn.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}
	if exists {
		return fmt.Errorf("database %s already exists", dbName)
	}

	// Create database using template
	templateDB := s.cfg.MultiTenant.TemplateDBName

	// Force disconnect all users from template database to allow cloning
	_, _ = masterConn.ExecContext(ctx, `
		SELECT pg_terminate_backend(pid) 
		FROM pg_stat_activity 
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`, templateDB)
	_, err = masterConn.ExecContext(ctx,
		fmt.Sprintf("CREATE DATABASE \"%s\" WITH TEMPLATE \"%s\"",
			sanitizeDBName(dbName),
			sanitizeDBName(templateDB)))
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Connect to the new database to clear copied data (Schema Only)
	tenantDB, err := database.ConnectTenant(&s.cfg.Database, dbName)
	if err != nil {
		return fmt.Errorf("failed to connect to new tenant database for cleanup: %w", err)
	}
	// defer tenantDB.Close() - Do not close cached connection

	// Truncate all tables to ensure empty schema
	_, err = tenantDB.ExecContext(ctx, `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename NOT IN ('dashboard_widgets')) LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	if err != nil {
		return fmt.Errorf("failed to clear data from new database: %w", err)
	}

	return nil
}

// CreateTenantAdmin creates the first admin user for a new tenant
func (s *TenantService) CreateTenantAdmin(ctx context.Context, org *models.Organization, req *models.CreateTenantAdminRequest) (*models.CreateTenantAdminResponse, error) {
	// Generate random 12-character password
	password := generateRandomPassword(12)

	// Hash password
	hashedPassword, err := security.HashPassword(password, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Connect to tenant database
	tenantDB, err := database.ConnectTenant(&s.cfg.Database, org.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to tenant database: %w", err)
	}
	// defer tenantDB.Close() - Do not close cached connection

	// Use a fresh context with timeout to avoid premature cancellation
	execCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create admin user
	userID := uuid.New()
	_, err = tenantDB.ExecContext(execCtx, `
		INSERT INTO users (
			id, username, email, password_hash, full_name, full_name_ar,
			role, is_active, two_factor_enabled, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, true, false, $8, $9)
	`, userID, req.Username, req.Email, hashedPassword, req.FullName, req.FullNameAr,
		models.RoleAdmin, time.Now(), time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant admin: %w", err)
	}

	return &models.CreateTenantAdminResponse{
		Username:          req.Username,
		Email:             req.Email,
		TemporaryPassword: password,
		OrganizationName:  org.Name,
		LoginURL:          fmt.Sprintf("https://%s/login", org.PrimaryDomain),
	}, nil
}

// GetOrganizationByID retrieves an organization by ID
func (s *TenantService) GetOrganizationByID(ctx context.Context, id uuid.UUID) (*models.Organization, error) {
	var org models.Organization
	var nameAr, contactPhone, logoURL sql.NullString
	var subscriptionPlanID sql.NullString
	var subscriptionStartDate, subscriptionExpires, verifiedAt, suspendedAt sql.NullTime

	err := s.masterDB.QueryRowContext(ctx, `
		SELECT id, name, name_ar, slug, primary_domain, database_name,
			   contact_email, contact_phone, status, subscription_plan_id,
			   subscription_status, subscription_start_date, subscription_expires_at, logo_url,
			   created_at, updated_at, verified_at, suspended_at, created_by
		FROM organizations WHERE id = $1
	`, id).Scan(
		&org.ID, &org.Name, &nameAr, &org.Slug, &org.PrimaryDomain, &org.DatabaseName,
		&org.ContactEmail, &contactPhone, &org.Status, &subscriptionPlanID,
		&org.SubscriptionStatus, &subscriptionStartDate, &subscriptionExpires, &logoURL,
		&org.CreatedAt, &org.UpdatedAt, &verifiedAt, &suspendedAt, &org.CreatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	// Handle nullable fields
	if nameAr.Valid {
		org.NameAr = &nameAr.String
	}
	if contactPhone.Valid {
		org.ContactPhone = &contactPhone.String
	}
	if logoURL.Valid {
		org.LogoURL = &logoURL.String
	}
	if subscriptionStartDate.Valid {
		org.SubscriptionStartDate = &subscriptionStartDate.Time
	}
	if subscriptionExpires.Valid {
		org.SubscriptionExpires = &subscriptionExpires.Time
	}
	if verifiedAt.Valid {
		org.VerifiedAt = &verifiedAt.Time
	}
	if suspendedAt.Valid {
		org.SuspendedAt = &suspendedAt.Time
	}
	if subscriptionPlanID.Valid {
		id, err := uuid.Parse(subscriptionPlanID.String)
		if err == nil {
			org.SubscriptionPlanID = &id
		}
	}

	return &org, nil
}

// UpdateOrganizationSubscription updates an organization's subscription details
func (s *TenantService) UpdateOrganizationSubscription(ctx context.Context, id uuid.UUID, req *models.UpdateOrganizationSubscriptionRequest) error {
	// Parse Plan ID
	planID, err := uuid.Parse(req.SubscriptionPlanID)
	if err != nil {
		return fmt.Errorf("invalid plan ID: %w", err)
	}

	// Verify plan exists
	var planExists bool
	err = s.masterDB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM subscription_plans WHERE id = $1)", planID).Scan(&planExists)
	if err != nil {
		return fmt.Errorf("failed to verify plan: %w", err)
	}
	if !planExists {
		return fmt.Errorf("subscription plan not found")
	}

	// Build dynamic query
	query := "UPDATE organizations SET subscription_plan_id = $1, updated_at = NOW()"
	args := []interface{}{planID}
	argIdx := 2

	if req.SubscriptionStart != nil {
		query += fmt.Sprintf(", subscription_start_date = $%d", argIdx)
		args = append(args, req.SubscriptionStart)
		argIdx++
	}

	if req.SubscriptionEnd != nil {
		query += fmt.Sprintf(", subscription_expires_at = $%d", argIdx)
		args = append(args, req.SubscriptionEnd)
		argIdx++
	}

	if req.SubscriptionStatus != "" {
		query += fmt.Sprintf(", subscription_status = $%d", argIdx)
		args = append(args, req.SubscriptionStatus)
		argIdx++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argIdx)
	args = append(args, id)

	result, err := s.masterDB.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

// GetOrganizationByDomain retrieves an organization by domain
func (s *TenantService) GetOrganizationByDomain(ctx context.Context, domain string) (*models.Organization, error) {
	var orgID uuid.UUID
	err := s.masterDB.QueryRowContext(ctx, `
		SELECT organization_id FROM allowed_domains 
		WHERE domain = $1 AND is_active = true AND verified = true
	`, domain).Scan(&orgID)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("domain not found or not verified")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization by domain: %w", err)
	}

	return s.GetOrganizationByID(ctx, orgID)
}

// ListOrganizations lists all organizations for SSA dashboard
func (s *TenantService) ListOrganizations(ctx context.Context, status string, limit, offset int) ([]models.OrganizationListItem, int, error) {
	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM organizations"
	if status != "" {
		countQuery += " WHERE status = $1"
		err := s.masterDB.QueryRowContext(ctx, countQuery, status).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
		}
	} else {
		err := s.masterDB.QueryRowContext(ctx, countQuery).Scan(&total)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to count organizations: %w", err)
		}
	}

	// Get organizations
	query := `
		SELECT id, name, name_ar, primary_domain, status, created_at
		FROM organizations
	`
	args := []any{}
	if status != "" {
		query += " WHERE status = $1"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args)+1) + " OFFSET $" + fmt.Sprint(len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.masterDB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	orgs := make([]models.OrganizationListItem, 0)
	for rows.Next() {
		var item models.OrganizationListItem
		var nameAr sql.NullString
		if err := rows.Scan(&item.ID, &item.Name, &nameAr, &item.PrimaryDomain, &item.Status, &item.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan organization: %w", err)
		}
		if nameAr.Valid {
			item.NameAr = &nameAr.String
		}
		orgs = append(orgs, item)
	}

	return orgs, total, nil
}

// SuspendOrganization suspends an organization
func (s *TenantService) SuspendOrganization(ctx context.Context, orgID uuid.UUID, reason string) error {
	now := time.Now()
	_, err := s.masterDB.ExecContext(ctx, `
		UPDATE organizations 
		SET status = $1, suspended_at = $2, updated_at = $3
		WHERE id = $4
	`, models.OrgStatusSuspended, now, now, orgID)
	if err != nil {
		return fmt.Errorf("failed to suspend organization: %w", err)
	}
	return nil
}

// ActivateOrganization activates a suspended organization
func (s *TenantService) ActivateOrganization(ctx context.Context, orgID uuid.UUID) error {
	now := time.Now()
	_, err := s.masterDB.ExecContext(ctx, `
		UPDATE organizations 
		SET status = $1, suspended_at = NULL, updated_at = $2
		WHERE id = $3
	`, models.OrgStatusActive, now, orgID)
	if err != nil {
		return fmt.Errorf("failed to activate organization: %w", err)
	}
	return nil
}

// ListSubscriptionPlans lists all available subscription plans
func (s *TenantService) ListSubscriptionPlans(ctx context.Context) ([]models.SubscriptionPlan, error) {
	rows, err := s.masterDB.QueryContext(ctx, `
		SELECT id, name, name_ar, code, price_monthly, price_yearly, currency, 
		       max_users, max_storage_gb, features, is_active, is_default, 
		       sort_order, description, description_ar
		FROM subscription_plans
		ORDER BY sort_order ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to list plans: %w", err)
	}
	defer rows.Close()

	var plans []models.SubscriptionPlan
	for rows.Next() {
		var p models.SubscriptionPlan
		var featuresJSON []byte

		if err := rows.Scan(
			&p.ID, &p.Name, &p.NameAr, &p.Code, &p.PriceMonthly, &p.PriceYearly, &p.Currency,
			&p.MaxUsers, &p.MaxStorageGB, &featuresJSON, &p.IsActive, &p.IsDefault,
			&p.SortOrder, &p.Description, &p.DescriptionAr,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plan: %w", err)
		}

		if err := json.Unmarshal(featuresJSON, &p.Features); err != nil {
			// Log error but continue? Or just default to empty
			p.Features = []string{}
		}

		plans = append(plans, p)
	}

	return plans, nil
}

// CreateSubscriptionPlan creates a new subscription plan
func (s *TenantService) CreateSubscriptionPlan(ctx context.Context, req *models.CreateSubscriptionPlanRequest) (*models.SubscriptionPlan, error) {
	// Check if code exists
	var exists bool
	err := s.masterDB.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM subscription_plans WHERE code = $1)", req.Code).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check plan code: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("subscription plan with code %s already exists", req.Code)
	}

	featuresJSON, err := json.Marshal(req.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal features: %w", err)
	}

	plan := &models.SubscriptionPlan{
		ID:            uuid.New(),
		Name:          req.Name,
		NameAr:        req.NameAr,
		Code:          req.Code,
		PriceMonthly:  req.PriceMonthly,
		PriceYearly:   req.PriceYearly,
		Currency:      req.Currency,
		MaxUsers:      req.MaxUsers,
		MaxStorageGB:  req.MaxStorageGB,
		Features:      req.Features,
		IsActive:      req.IsActive,
		IsDefault:     req.IsDefault,
		SortOrder:     req.SortOrder,
		Description:   req.Description,
		DescriptionAr: req.DescriptionAr,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	_, err = s.masterDB.ExecContext(ctx, `
		INSERT INTO subscription_plans (
			id, name, name_ar, code, price_monthly, price_yearly, currency,
			max_users, max_storage_gb, features, is_active, is_default,
			sort_order, description, description_ar, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`, plan.ID, plan.Name, plan.NameAr, plan.Code, plan.PriceMonthly, plan.PriceYearly, plan.Currency,
		plan.MaxUsers, plan.MaxStorageGB, featuresJSON, plan.IsActive, plan.IsDefault,
		plan.SortOrder, plan.Description, plan.DescriptionAr, plan.CreatedAt, plan.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create subscription plan: %w", err)
	}

	return plan, nil
}

// GetSubscriptionPlan retrieves a subscription plan by ID
func (s *TenantService) GetSubscriptionPlan(ctx context.Context, id uuid.UUID) (*models.SubscriptionPlan, error) {
	var plan models.SubscriptionPlan
	var featuresJSON []byte

	err := s.masterDB.QueryRowContext(ctx, `
		SELECT id, name, name_ar, code, price_monthly, price_yearly, currency,
			   max_users, max_storage_gb, features, is_active, is_default,
			   sort_order, description, description_ar, created_at, updated_at
		FROM subscription_plans WHERE id = $1
	`, id).Scan(
		&plan.ID, &plan.Name, &plan.NameAr, &plan.Code, &plan.PriceMonthly, &plan.PriceYearly, &plan.Currency,
		&plan.MaxUsers, &plan.MaxStorageGB, &featuresJSON, &plan.IsActive, &plan.IsDefault,
		&plan.SortOrder, &plan.Description, &plan.DescriptionAr, &plan.CreatedAt, &plan.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("subscription plan not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription plan: %w", err)
	}

	if err := json.Unmarshal(featuresJSON, &plan.Features); err != nil {
		plan.Features = []string{}
	}

	return &plan, nil
}

// UpdateSubscriptionPlan updates an existing subscription plan
func (s *TenantService) UpdateSubscriptionPlan(ctx context.Context, id uuid.UUID, req *models.UpdateSubscriptionPlanRequest) (*models.SubscriptionPlan, error) {
	// Check existence
	_, err := s.GetSubscriptionPlan(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build query dynamically
	query := "UPDATE subscription_plans SET updated_at = $1"
	args := []any{time.Now()}
	argID := 2

	if req.Name != nil {
		query += fmt.Sprintf(", name = $%d", argID)
		args = append(args, *req.Name)
		argID++
	}
	if req.NameAr != nil {
		query += fmt.Sprintf(", name_ar = $%d", argID)
		args = append(args, *req.NameAr)
		argID++
	}
	if req.PriceMonthly != nil {
		query += fmt.Sprintf(", price_monthly = $%d", argID)
		args = append(args, *req.PriceMonthly)
		argID++
	}
	if req.PriceYearly != nil {
		query += fmt.Sprintf(", price_yearly = $%d", argID)
		args = append(args, *req.PriceYearly)
		argID++
	}
	if req.MaxUsers != nil {
		query += fmt.Sprintf(", max_users = $%d", argID)
		args = append(args, *req.MaxUsers)
		argID++
	}
	if req.MaxStorageGB != nil {
		query += fmt.Sprintf(", max_storage_gb = $%d", argID)
		args = append(args, *req.MaxStorageGB)
		argID++
	}
	if req.Features != nil {
		featuresJSON, _ := json.Marshal(req.Features)
		query += fmt.Sprintf(", features = $%d", argID)
		args = append(args, featuresJSON)
		argID++
	}
	if req.IsActive != nil {
		query += fmt.Sprintf(", is_active = $%d", argID)
		args = append(args, *req.IsActive)
		argID++
	}
	if req.IsDefault != nil {
		query += fmt.Sprintf(", is_default = $%d", argID)
		args = append(args, *req.IsDefault)
		argID++
	}
	if req.SortOrder != nil {
		query += fmt.Sprintf(", sort_order = $%d", argID)
		args = append(args, *req.SortOrder)
		argID++
	}
	if req.Description != nil {
		query += fmt.Sprintf(", description = $%d", argID)
		args = append(args, *req.Description)
		argID++
	}
	if req.DescriptionAr != nil {
		query += fmt.Sprintf(", description_ar = $%d", argID)
		args = append(args, *req.DescriptionAr)
		argID++
	}

	query += fmt.Sprintf(" WHERE id = $%d", argID)
	args = append(args, id)

	_, err = s.masterDB.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription plan: %w", err)
	}

	return s.GetSubscriptionPlan(ctx, id)
}

// DeleteSubscriptionPlan deletes a subscription plan
func (s *TenantService) DeleteSubscriptionPlan(ctx context.Context, id uuid.UUID) error {
	// Check if any organizations are using this plan
	var count int
	err := s.masterDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM organizations WHERE subscription_plan_id = $1", id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check plan usage: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("cannot delete plan used by %d organizations", count)
	}

	res, err := s.masterDB.ExecContext(ctx, "DELETE FROM subscription_plans WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to delete plan: %w", err)
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("plan not found")
	}

	return nil
}

// GetOrganizationStats returns statistics for SSA dashboard
func (s *TenantService) GetOrganizationStats(ctx context.Context) (*models.OrganizationStats, error) {
	stats := &models.OrganizationStats{}

	err := s.masterDB.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN status = 'pending_verification' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'suspended' THEN 1 END) as suspended
		FROM organizations
	`).Scan(&stats.TotalOrganizations, &stats.ActiveOrganizations,
		&stats.PendingVerification, &stats.SuspendedOrganizations)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization stats: %w", err)
	}

	return stats, nil
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// generateSlug creates a URL-friendly slug from name
func generateSlug(name string) string {
	// Remove special characters
	reg := regexp.MustCompile(`[^a-zA-Z0-9\s-]`)
	slug := reg.ReplaceAllString(name, "")

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Lowercase
	slug = strings.ToLower(slug)

	// Remove consecutive hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")

	// Trim hyphens
	slug = strings.Trim(slug, "-")

	// Limit length
	if len(slug) > 50 {
		slug = slug[:50]
	}

	return slug
}

// generateRandomCode generates a random alphanumeric code
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// generateRandomPassword generates a secure random password
func generateRandomPassword(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*"
	b := make([]byte, length)
	rand.Read(b)
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b)
}

// hashVerificationCode hashes the verification code using SHA256
func hashVerificationCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

// sanitizeDBName sanitizes database name for SQL
func sanitizeDBName(name string) string {
	// Allow alphanumeric, underscores, and hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return reg.ReplaceAllString(name, "")
}

// ============================================================================
// ORGANIZATION DELETION
// ============================================================================

// DeletionRequestResult contains the result of a deletion request
type DeletionRequestResult struct {
	RequestID string    `json:"request_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message"`
}

// RequestOrganizationDeletion initiates the deletion process by generating OTP
func (s *TenantService) RequestOrganizationDeletion(ctx context.Context, orgID, adminID uuid.UUID, adminEmail, adminName string, requestIP string, emailService interface {
	SendDeletionOTPEmail(ctx context.Context, toEmail, adminName, orgName, code string) error
}) (*DeletionRequestResult, error) {
	// Get organization to verify it exists
	org, err := s.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	// Check if there's already a pending deletion request
	var existingCount int
	err = s.masterDB.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM org_deletion_requests WHERE organization_id = $1 AND expires_at > NOW()",
		orgID).Scan(&existingCount)
	if err == nil && existingCount > 0 {
		return nil, fmt.Errorf("طلب حذف معلق موجود بالفعل. يرجى الانتظار أو استخدام الرمز المرسل")
	}

	// Delete any expired requests
	_, _ = s.masterDB.ExecContext(ctx,
		"DELETE FROM org_deletion_requests WHERE organization_id = $1",
		orgID)

	// Generate 6-digit OTP
	code := generateRandomCode(6)
	hashedCode := hashVerificationCode(code)
	expiresAt := time.Now().Add(10 * time.Minute)
	requestID := uuid.New()

	// Insert deletion request
	_, err = s.masterDB.ExecContext(ctx, `
		INSERT INTO org_deletion_requests (
			id, organization_id, email, code_hash, expires_at, request_ip
		) VALUES ($1, $2, $3, $4, $5, $6)
	`, requestID, orgID, adminEmail, hashedCode, expiresAt, requestIP)
	if err != nil {
		return nil, fmt.Errorf("فشل في إنشاء طلب الحذف: %w", err)
	}

	// Send OTP via email
	if emailService != nil {
		err = emailService.SendDeletionOTPEmail(ctx, adminEmail, adminName, org.Name, code)
		if err != nil {
			// Rollback the request if email fails
			s.masterDB.ExecContext(ctx, "DELETE FROM org_deletion_requests WHERE id = $1", requestID)
			return nil, fmt.Errorf("فشل في إرسال رمز التحقق: %w", err)
		}
	}

	return &DeletionRequestResult{
		RequestID: requestID.String(),
		ExpiresAt: expiresAt,
		Message:   "تم إرسال رمز التحقق إلى بريدك الإلكتروني",
	}, nil
}

// ConfirmOrganizationDeletion verifies OTP and performs deletion
func (s *TenantService) ConfirmOrganizationDeletion(ctx context.Context, orgID uuid.UUID, code string, adminID uuid.UUID, verifyIP string) error {
	hashedCode := hashVerificationCode(code)

	// Get pending deletion request
	var request struct {
		ID          uuid.UUID
		CodeHash    string
		ExpiresAt   time.Time
		Attempts    int
		MaxAttempts int
	}
	err := s.masterDB.QueryRowContext(ctx, `
		SELECT id, code_hash, expires_at, attempts, max_attempts
		FROM org_deletion_requests
		WHERE organization_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
		LIMIT 1
	`, orgID).Scan(&request.ID, &request.CodeHash, &request.ExpiresAt, &request.Attempts, &request.MaxAttempts)

	if err != nil {
		return fmt.Errorf("لا يوجد طلب حذف معلق أو انتهت صلاحيته")
	}

	// Check expiry
	if time.Now().After(request.ExpiresAt) {
		s.masterDB.ExecContext(ctx, "DELETE FROM org_deletion_requests WHERE id = $1", request.ID)
		return fmt.Errorf("انتهت صلاحية رمز التحقق")
	}

	// Check max attempts
	if request.Attempts >= request.MaxAttempts {
		s.masterDB.ExecContext(ctx, "DELETE FROM org_deletion_requests WHERE id = $1", request.ID)
		return fmt.Errorf("تجاوزت الحد الأقصى للمحاولات")
	}

	// Increment attempts
	_, _ = s.masterDB.ExecContext(ctx,
		"UPDATE org_deletion_requests SET attempts = attempts + 1 WHERE id = $1",
		request.ID)

	// Verify code
	if request.CodeHash != hashedCode {
		remainingAttempts := request.MaxAttempts - request.Attempts - 1
		return fmt.Errorf("رمز التحقق غير صحيح. المحاولات المتبقية: %d", remainingAttempts)
	}

	// Get organization details for deletion
	org, err := s.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("فشل في جلب بيانات المنظمة: %w", err)
	}

	// Perform the actual deletion
	err = s.DeleteOrganization(ctx, org, adminID, verifyIP)
	if err != nil {
		return err
	}

	// Clean up deletion request
	s.masterDB.ExecContext(ctx, "DELETE FROM org_deletion_requests WHERE id = $1", request.ID)

	return nil
}

// DeleteOrganization performs the actual deletion of organization and its database
func (s *TenantService) DeleteOrganization(ctx context.Context, org *models.Organization, deletedBy uuid.UUID, deleteIP string) error {
	// Start transaction for master DB operations
	tx, err := s.masterDB.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("فشل في بدء عملية الحذف: %w", err)
	}
	defer tx.Rollback()

	// Log the deletion in audit before deleting
	_, err = tx.ExecContext(ctx, `
		INSERT INTO super_admin_audit (admin_id, admin_email, action, resource_type, resource_id, old_values, ip_address)
		VALUES ($1, '', 'delete_organization', 'organization', $2, $3, $4)
	`, deletedBy, org.ID, fmt.Sprintf(`{"name":"%s","database":"%s","domain":"%s"}`, org.Name, org.DatabaseName, org.PrimaryDomain), deleteIP)
	if err != nil {
		return fmt.Errorf("فشل في تسجيل العملية في سجل المراجعة: %w", err)
	}

	// Delete from organizations table (CASCADE will handle related tables)
	result, err := tx.ExecContext(ctx, "DELETE FROM organizations WHERE id = $1", org.ID)
	if err != nil {
		return fmt.Errorf("فشل في حذف سجل المنظمة: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("المنظمة غير موجودة")
	}

	// Commit master DB transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("فشل في تأكيد عملية الحذف: %w", err)
	}

	// Drop the tenant database
	// Note: This is done after committing the transaction to ensure master records are deleted first
	err = s.DropTenantDatabase(ctx, org.DatabaseName)
	if err != nil {
		// Log but don't fail - the organization record is already deleted
		// This prevents orphaned organization records
		fmt.Printf("Warning: Failed to drop database %s: %v\n", org.DatabaseName, err)
	}

	return nil
}

// DropTenantDatabase drops the tenant's database
func (s *TenantService) DropTenantDatabase(ctx context.Context, dbName string) error {
	// Sanitize database name to prevent SQL injection
	safeDBName := sanitizeDBName(dbName)
	if safeDBName == "" {
		return fmt.Errorf("invalid database name")
	}

	// Terminate all connections to the database first
	_, err := s.masterDB.ExecContext(ctx, `
		SELECT pg_terminate_backend(pid) 
		FROM pg_stat_activity 
		WHERE datname = $1 AND pid <> pg_backend_pid()
	`, safeDBName)
	if err != nil {
		// Don't fail if this errors - database might not have any connections
		fmt.Printf("Warning: Could not terminate connections to %s: %v\n", safeDBName, err)
	}

	// Check if database exists before trying to drop
	var exists bool
	err = s.masterDB.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)",
		safeDBName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check database existence: %w", err)
	}

	if !exists {
		// Database doesn't exist, nothing to drop
		return nil
	}

	// Drop the database
	// Note: DROP DATABASE cannot be run inside a transaction
	_, err = s.masterDB.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE IF EXISTS "%s"`, safeDBName))
	if err != nil {
		return fmt.Errorf("فشل في حذف قاعدة البيانات: %w", err)
	}

	return nil
}
