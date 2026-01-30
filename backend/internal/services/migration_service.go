package services

import (
	"context"
	"cso-erp/internal/database"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MigrationService handles database migrations for all tenants
type MigrationService struct {
	masterDB *database.DB
	cfg      struct {
		DBHost     string
		DBPort     string
		DBUser     string
		DBPassword string
		DBSSLMode  string
	}
}

// MigrationResult represents the result of a migration run
type MigrationResult struct {
	DatabaseName  string    `json:"database_name"`
	Status        string    `json:"status"` // success, failed, skipped
	MigrationsRun []string  `json:"migrations_run,omitempty"`
	Error         string    `json:"error,omitempty"`
	Duration      string    `json:"duration"`
	CompletedAt   time.Time `json:"completed_at"`
}

// MigrationRunResult represents the overall result of running migrations
type MigrationRunResult struct {
	Success          bool              `json:"success"`
	TotalDatabases   int               `json:"total_databases"`
	UpdatedDatabases int               `json:"updated_databases"`
	FailedDatabases  int               `json:"failed_databases"`
	TemplateUpdated  bool              `json:"template_updated"`
	Results          []MigrationResult `json:"results"`
	StartedAt        time.Time         `json:"started_at"`
	CompletedAt      time.Time         `json:"completed_at"`
	TotalDuration    string            `json:"total_duration"`
}

func NewMigrationService(masterDB *database.DB) *MigrationService {
	return &MigrationService{
		masterDB: masterDB,
		cfg: struct {
			DBHost     string
			DBPort     string
			DBUser     string
			DBPassword string
			DBSSLMode  string
		}{
			DBHost:     getEnvOrDefault("DB_HOST", "localhost"),
			DBPort:     getEnvOrDefault("DB_PORT", "5432"),
			DBUser:     getEnvOrDefault("DB_USER", "postgres"),
			DBPassword: getEnvOrDefault("DB_PASSWORD", ""),
			DBSSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetPendingMigrations returns list of migrations that haven't been applied to a database
func (s *MigrationService) GetPendingMigrations(ctx context.Context, dbName string) ([]string, error) {
	// Get all migration files
	migrationsDir := "database/migrations"
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	var allMigrations []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") && strings.HasPrefix(f.Name(), "0") {
			// Only include numbered migrations (e.g., 001_xxx.sql, 011_xxx.sql)
			migrationName := strings.TrimSuffix(f.Name(), ".sql")
			allMigrations = append(allMigrations, migrationName)
		}
	}
	sort.Strings(allMigrations)

	// Connect to target database
	db, err := s.connectToDatabase(dbName)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	// Check which migrations are already applied
	var appliedMigrations []string
	rows, err := db.QueryContext(ctx, `
		SELECT migration_name FROM applied_migrations ORDER BY migration_name
	`)
	if err != nil {
		// Table might not exist yet, return all migrations
		return allMigrations, nil
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		appliedMigrations = append(appliedMigrations, name)
	}

	// Filter out already applied migrations
	appliedMap := make(map[string]bool)
	for _, m := range appliedMigrations {
		appliedMap[m] = true
	}

	var pending []string
	for _, m := range allMigrations {
		if !appliedMap[m] {
			pending = append(pending, m)
		}
	}

	return pending, nil
}

// RunMigrations runs all pending migrations on all tenant databases
func (s *MigrationService) RunMigrations(ctx context.Context, adminID uuid.UUID) (*MigrationRunResult, error) {
	startTime := time.Now()
	result := &MigrationRunResult{
		StartedAt: startTime,
		Results:   []MigrationResult{},
	}

	// Get all active organizations from master DB
	rows, err := s.masterDB.QueryContext(ctx, `
		SELECT database_name FROM organizations WHERE status = 'active'
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var dbName string
		if err := rows.Scan(&dbName); err != nil {
			continue
		}
		databases = append(databases, dbName)
	}

	// Add template database
	databases = append([]string{"cso-erp"}, databases...)
	result.TotalDatabases = len(databases)

	// Run migrations on each database
	for _, dbName := range databases {
		migResult := s.runMigrationOnDatabase(ctx, dbName)
		result.Results = append(result.Results, migResult)

		if migResult.Status == "success" {
			result.UpdatedDatabases++
			if dbName == "cso-erp" {
				result.TemplateUpdated = true
			}
		} else if migResult.Status == "failed" {
			result.FailedDatabases++
		}
	}

	result.CompletedAt = time.Now()
	result.TotalDuration = result.CompletedAt.Sub(startTime).String()
	result.Success = result.FailedDatabases == 0

	return result, nil
}

func (s *MigrationService) runMigrationOnDatabase(ctx context.Context, dbName string) MigrationResult {
	startTime := time.Now()
	result := MigrationResult{
		DatabaseName: dbName,
		CompletedAt:  time.Now(),
	}

	// Get pending migrations
	pending, err := s.GetPendingMigrations(ctx, dbName)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.Duration = time.Since(startTime).String()
		return result
	}

	if len(pending) == 0 {
		result.Status = "skipped"
		result.Duration = time.Since(startTime).String()
		return result
	}

	// Connect to database
	db, err := s.connectToDatabase(dbName)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
		result.Duration = time.Since(startTime).String()
		return result
	}
	defer db.Close()

	// Run each pending migration
	migrationsDir := "database/migrations"
	for _, migrationName := range pending {
		migrationFile := filepath.Join(migrationsDir, migrationName+".sql")
		content, err := os.ReadFile(migrationFile)
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("failed to read %s: %v", migrationName, err)
			result.Duration = time.Since(startTime).String()
			return result
		}

		// Execute migration
		_, err = db.ExecContext(ctx, string(content))
		if err != nil {
			result.Status = "failed"
			result.Error = fmt.Sprintf("failed to run %s: %v", migrationName, err)
			result.Duration = time.Since(startTime).String()
			return result
		}

		result.MigrationsRun = append(result.MigrationsRun, migrationName)
	}

	result.Status = "success"
	result.Duration = time.Since(startTime).String()
	result.CompletedAt = time.Now()
	return result
}

func (s *MigrationService) connectToDatabase(dbName string) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		s.cfg.DBHost, s.cfg.DBPort, s.cfg.DBUser, s.cfg.DBPassword, dbName, s.cfg.DBSSLMode,
	)
	return sql.Open("postgres", connStr)
}

// GetMigrationStatus returns the status of migrations for all databases
func (s *MigrationService) GetMigrationStatus(ctx context.Context) ([]map[string]interface{}, error) {
	var status []map[string]interface{}

	// Get all organizations
	rows, err := s.masterDB.QueryContext(ctx, `
		SELECT database_name, name FROM organizations WHERE status = 'active'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	databases := []struct {
		DBName string
		Name   string
	}{
		{DBName: "cso-erp", Name: "Template Database"},
	}

	for rows.Next() {
		var dbName, name string
		if err := rows.Scan(&dbName, &name); err != nil {
			continue
		}
		databases = append(databases, struct {
			DBName string
			Name   string
		}{DBName: dbName, Name: name})
	}

	for _, db := range databases {
		pending, _ := s.GetPendingMigrations(ctx, db.DBName)
		status = append(status, map[string]interface{}{
			"database_name":      db.DBName,
			"organization_name":  db.Name,
			"pending_migrations": len(pending),
			"pending_list":       pending,
		})
	}

	return status, nil
}
