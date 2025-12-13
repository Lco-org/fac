package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

	org := &models.Organization{
		ID:                 uuid.New(),
		Name:               req.Name,
		NameAr:             req.NameAr,
		Slug:               slug,
		PrimaryDomain:      req.Domain,
		DatabaseName:       dbName,
		ContactEmail:       req.ContactEmail,
		ContactPhone:       req.ContactPhone,
		Status:             models.OrgStatusPendingVerification,
		SubscriptionStatus: "trial",
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
		CreatedBy:          createdBy,
	}

	// Insert organization
	_, err = s.masterDB.ExecContext(ctx, `
		INSERT INTO organizations (
			id, name, name_ar, slug, primary_domain, database_name,
			contact_email, contact_phone, status, subscription_status,
			created_at, updated_at, created_by
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, org.ID, org.Name, org.NameAr, org.Slug, org.PrimaryDomain, org.DatabaseName,
		org.ContactEmail, org.ContactPhone, org.Status, org.SubscriptionStatus,
		org.CreatedAt, org.UpdatedAt, org.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Insert into allowed_domains
	_, err = s.masterDB.ExecContext(ctx, `
		INSERT INTO allowed_domains (id, organization_id, domain, is_primary, is_active)
		VALUES ($1, $2, $3, true, true)
	`, uuid.New(), org.ID, req.Domain)
	if err != nil {
		return nil, fmt.Errorf("failed to add domain: %w", err)
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
			id, organization_id, email, hash_code, expires_at, request_ip
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
		SELECT id, organization_id, email, hash_code, expires_at, verified, attempts, max_attempts
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
	defer tenantDB.Close()

	// Truncate all tables to ensure empty schema
	_, err = tenantDB.ExecContext(ctx, `
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
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

	// Create admin user
	userID := uuid.New()
	_, err = tenantDB.ExecContext(ctx, `
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
	var subscriptionExpires, verifiedAt, suspendedAt sql.NullTime

	err := s.masterDB.QueryRowContext(ctx, `
		SELECT id, name, name_ar, slug, primary_domain, database_name,
			   contact_email, contact_phone, status, subscription_plan_id,
			   subscription_status, subscription_expires_at, logo_url,
			   created_at, updated_at, verified_at, suspended_at, created_by
		FROM organizations WHERE id = $1
	`, id).Scan(
		&org.ID, &org.Name, &nameAr, &org.Slug, &org.PrimaryDomain, &org.DatabaseName,
		&org.ContactEmail, &contactPhone, &org.Status, &subscriptionPlanID,
		&org.SubscriptionStatus, &subscriptionExpires, &logoURL,
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
	if subscriptionExpires.Valid {
		org.SubscriptionExpires = &subscriptionExpires.Time
	}
	if verifiedAt.Valid {
		org.VerifiedAt = &verifiedAt.Time
	}
	if suspendedAt.Valid {
		org.SuspendedAt = &suspendedAt.Time
	}

	return &org, nil
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

// hashVerificationCode hashes the verification code
func hashVerificationCode(code string) string {
	hash := make([]byte, 32)
	copy(hash, []byte(code))
	return hex.EncodeToString(hash)
}

// sanitizeDBName sanitizes database name for SQL
func sanitizeDBName(name string) string {
	// Allow alphanumeric, underscores, and hyphens
	reg := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	return reg.ReplaceAllString(name, "")
}
