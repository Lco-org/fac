package middleware

import (
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// TenantContext holds tenant information for the current request
type TenantContext struct {
	OrganizationID   string
	OrganizationSlug string
	DatabaseName     string
	PrimaryDomain    string
}

// DBHolder wraps *database.DB WITHOUT implementing io.Closer
// This prevents Fiber from auto-closing the DB when request ends
type DBHolder struct {
	DB *database.DB
}

// TenantResolver middleware resolves tenant from request subdomain/domain
type TenantResolver struct {
	cfg           *config.Config
	masterDB      *sql.DB
	tenantDBCache map[string]*database.DB
	cacheMutex    sync.RWMutex
}

// NewTenantResolver creates a new tenant resolver
func NewTenantResolver(cfg *config.Config, masterDB *sql.DB) *TenantResolver {
	return &TenantResolver{
		cfg:           cfg,
		masterDB:      masterDB,
		tenantDBCache: make(map[string]*database.DB),
	}
}

// ResolveTenant middleware extracts tenant from request and injects tenant context
func (tr *TenantResolver) ResolveTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Extract domain from request
		host := c.Hostname()

		// Remove port if present
		if colonIdx := strings.Index(host, ":"); colonIdx != -1 {
			host = host[:colonIdx]
		}

		// Skip tenant resolution for certain paths
		path := c.Path()

		// DEBUG: Log every request
		log.Printf("[TenantResolver] Request: %s %s from host: %s", c.Method(), path, host)

		if tr.shouldSkipResolution(path) {
			log.Printf("[TenantResolver] Skipping resolution for path: %s", path)
			return c.Next()
		}

		// For localhost/development, check header or use default 'cso-erp'
		if host == "localhost" || host == "127.0.0.1" || strings.HasPrefix(host, "192.168.") {
			// Check for X-Tenant-Slug header (development/testing)
			tenantSlug := c.Get("X-Tenant-Slug")
			if tenantSlug != "" {
				return tr.resolveBySlug(c, tenantSlug)
			}

			// Priority 1: Check for JWT tenant_db claim (for authenticated users)
			// This is set by parsing the JWT token cookie/header
			tenantDBFromJWT := tr.extractTenantFromJWT(c)
			if tenantDBFromJWT != "" {
				log.Printf("[TenantResolver] Using tenant DB from JWT: %s", tenantDBFromJWT)
				c.Locals("tenant", &TenantContext{
					OrganizationSlug: "jwt",
					DatabaseName:     tenantDBFromJWT,
				})
				c.Locals("tenantDB", &DBHolder{DB: tr.getTenantDB(tenantDBFromJWT)})
				return c.Next()
			}

			// Default localhost to cso-erp (Template/Master DB acting as default tenant)
			// This fixes "Database connection not found" when running locally without a tenant
			log.Printf("[TenantResolver] Fallback to default: cso-erp")
			tenantDB := tr.getTenantDB("cso-erp")
			if tenantDB == nil {
				log.Printf("[TenantResolver] ERROR: Failed to get cso-erp connection!")
			}
			c.Locals("tenant", &TenantContext{
				OrganizationSlug: "default",
				DatabaseName:     "cso-erp", // Use the main DB for localhost
			})
			c.Locals("tenantDB", &DBHolder{DB: tenantDB})
			return c.Next()
		}

		// Production: resolve by subdomain or full domain
		return tr.resolveByDomain(c, host)
	}
}

// shouldSkipResolution returns true if path should skip tenant resolution
func (tr *TenantResolver) shouldSkipResolution(path string) bool {
	skipPaths := []string{
		"/health",
		"/api/v1/ssa",         // SSA routes use master DB
		"/api/v1/system-auth", // Public SSA auth
		"/b8fb17",             // Hidden SSA panel
	}

	for _, skip := range skipPaths {
		if strings.HasPrefix(path, skip) {
			return true
		}
	}
	return false
}

// resolveBySlug resolves tenant by slug (development mode)
func (tr *TenantResolver) resolveBySlug(c *fiber.Ctx, slug string) error {
	tenant, err := tr.getTenantBySlug(slug)
	if err != nil {
		log.Printf("Failed to resolve tenant by slug '%s': %v", slug, err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Tenant not found",
		})
	}

	if tenant == nil {
		// No tenant found, continue without tenant context (use default DB)
		c.Locals("tenant", nil)
		return c.Next()
	}

	// Inject tenant context
	c.Locals("tenant", tenant)
	c.Locals("tenantDB", &DBHolder{DB: tr.getTenantDB(tenant.DatabaseName)})

	return c.Next()
}

// resolveByDomain resolves tenant by domain/subdomain
func (tr *TenantResolver) resolveByDomain(c *fiber.Ctx, host string) error {
	// First, try to find by exact domain match
	tenant, err := tr.getTenantByDomain(host)
	if err != nil {
		log.Printf("Failed to resolve tenant by domain '%s': %v", host, err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "خطأ في تحديد المنظمة",
		})
	}

	if tenant == nil {
		// Try subdomain extraction (e.g., tenant1.erp-platform.com)
		parts := strings.Split(host, ".")
		if len(parts) >= 3 {
			subdomain := parts[0]
			tenant, err = tr.getTenantBySlug(subdomain)
			if err != nil {
				log.Printf("Failed to resolve tenant by subdomain '%s': %v", subdomain, err)
			}
		}
	}

	if tenant == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "المنظمة غير موجودة",
			"host":  host,
		})
	}

	// Check if organization is active
	if tenant.OrganizationID == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "المنظمة غير متاحة",
		})
	}

	// Inject tenant context
	c.Locals("tenant", tenant)

	// Get tenant database connection
	tenantDB := tr.getTenantDB(tenant.DatabaseName)
	c.Locals("tenantDB", &DBHolder{DB: tenantDB})

	return c.Next()
}

// getTenantBySlug looks up tenant by slug from master database
func (tr *TenantResolver) getTenantBySlug(slug string) (*TenantContext, error) {
	query := `
		SELECT id, slug, database_name, primary_domain
		FROM organizations
		WHERE slug = $1 AND status = 'active'
	`

	var tenant TenantContext
	err := tr.masterDB.QueryRow(query, slug).Scan(
		&tenant.OrganizationID,
		&tenant.OrganizationSlug,
		&tenant.DatabaseName,
		&tenant.PrimaryDomain,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// getTenantByDomain looks up tenant by domain from master database
func (tr *TenantResolver) getTenantByDomain(domain string) (*TenantContext, error) {
	query := `
		SELECT o.id, o.slug, o.database_name, o.primary_domain
		FROM organizations o
		LEFT JOIN allowed_domains ad ON o.id = ad.organization_id
		WHERE (o.primary_domain = $1 OR ad.domain = $1) AND o.status = 'active'
		LIMIT 1
	`

	var tenant TenantContext
	err := tr.masterDB.QueryRow(query, domain).Scan(
		&tenant.OrganizationID,
		&tenant.OrganizationSlug,
		&tenant.DatabaseName,
		&tenant.PrimaryDomain,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &tenant, nil
}

// getTenantDB gets or creates a database connection for the tenant
func (tr *TenantResolver) getTenantDB(dbName string) *database.DB {
	// Check cache first
	tr.cacheMutex.RLock()
	if db, exists := tr.tenantDBCache[dbName]; exists {
		tr.cacheMutex.RUnlock()
		return db
	}
	tr.cacheMutex.RUnlock()

	// Create new connection
	tr.cacheMutex.Lock()
	defer tr.cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if db, exists := tr.tenantDBCache[dbName]; exists {
		return db
	}

	// Create database config for tenant
	tenantDBConfig := &config.DatabaseConfig{
		Host:     tr.cfg.Database.Host,
		Port:     tr.cfg.Database.Port,
		User:     tr.cfg.Database.User,
		Password: tr.cfg.Database.Password,
		DBName:   dbName,
		SSLMode:  tr.cfg.Database.SSLMode,
	}

	db, err := database.Connect(tenantDBConfig)
	if err != nil {
		log.Printf("Failed to connect to tenant database '%s': %v", dbName, err)
		return nil
	}

	tr.tenantDBCache[dbName] = db
	log.Printf("Connected to tenant database: %s", dbName)

	return db
}

// CloseAllConnections closes all cached tenant database connections
func (tr *TenantResolver) CloseAllConnections() {
	tr.cacheMutex.Lock()
	defer tr.cacheMutex.Unlock()

	for name, db := range tr.tenantDBCache {
		if db != nil {
			db.Close()
			log.Printf("Closed tenant database connection: %s", name)
		}
	}
	tr.tenantDBCache = make(map[string]*database.DB)
}

// GetTenant extracts tenant context from fiber context
func GetTenant(c *fiber.Ctx) *TenantContext {
	tenant := c.Locals("tenant")
	if tenant == nil {
		return nil
	}
	return tenant.(*TenantContext)
}

// GetTenantDB extracts tenant database from fiber context
func GetTenantDB(c *fiber.Ctx) *database.DB {
	holder := c.Locals("tenantDB")
	if holder == nil {
		return nil
	}
	return holder.(*DBHolder).DB
}

// RequireTenant middleware ensures tenant context exists
func RequireTenant() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant := GetTenant(c)
		if tenant == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Tenant context required",
			})
		}
		return c.Next()
	}
}

// TenantInfo returns tenant info for debugging
func (tr *TenantResolver) TenantInfo() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tenant := GetTenant(c)
		if tenant == nil {
			return c.JSON(fiber.Map{
				"tenant": nil,
				"host":   c.Hostname(),
			})
		}
		return c.JSON(fiber.Map{
			"tenant":         tenant.OrganizationSlug,
			"database":       tenant.DatabaseName,
			"primary_domain": tenant.PrimaryDomain,
		})
	}
}

// GetCachedConnectionCount returns the number of cached connections
func (tr *TenantResolver) GetCachedConnectionCount() int {
	tr.cacheMutex.RLock()
	defer tr.cacheMutex.RUnlock()
	return len(tr.tenantDBCache)
}

// String representation for logging
func (t *TenantContext) String() string {
	return fmt.Sprintf("Tenant{slug: %s, db: %s}", t.OrganizationSlug, t.DatabaseName)
}

// extractTenantFromJWT extracts tenant_db claim from JWT token in cookie or header
func (tr *TenantResolver) extractTenantFromJWT(c *fiber.Ctx) string {
	// Try cookie first
	tokenString := c.Cookies("auth_token")

	// Try Authorization header if no cookie
	if tokenString == "" {
		authHeader := c.Get("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}
	}

	if tokenString == "" {
		return ""
	}

	// Parse JWT token (without validation - just extracting claims)
	// We don't validate here because AuthMiddleware will do that
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return ""
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if tenantDB, exists := claims["tenant_db"].(string); exists {
			return tenantDB
		}
	}

	return ""
}
