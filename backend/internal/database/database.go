package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"cso-erp/internal/config"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

var dbInstance *DB

// Connect establishes database connection
func Connect(cfg *config.DatabaseConfig) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	dbInstance = &DB{db}
	log.Println("✅ Database connected successfully")
	return dbInstance, nil
}

// GetDB returns the database instance
func GetDB() *DB {
	return dbInstance
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return db.DB.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
}

// WithTransaction executes a function within a transaction
func (db *DB) WithTransaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.DB != nil {
		log.Println("Closing database connection...")
		// DEBUG: Trace who is closing the DB (commented out after fix)
		// debug.PrintStack()
		return db.DB.Close()
	}
	return nil
}

// ============================================================================
// MULTI-TENANT CONNECTION SUPPORT
// ============================================================================

// tenantConnections holds connections to tenant databases
var tenantConnections = make(map[string]*DB)
var tenantMu sync.RWMutex

// ConnectTenant connects to a specific tenant database
func ConnectTenant(cfg *config.DatabaseConfig, dbName string) (*DB, error) {
	// Check if connection already exists
	tenantMu.RLock()
	if db, exists := tenantConnections[dbName]; exists {
		tenantMu.RUnlock()
		return db, nil
	}
	tenantMu.RUnlock()

	// Create new connection
	tenantCfg := &config.DatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		DBName:   dbName,
		SSLMode:  cfg.SSLMode,
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		tenantCfg.Host, tenantCfg.Port, tenantCfg.User, tenantCfg.Password, tenantCfg.DBName, tenantCfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open tenant database %s: %w", dbName, err)
	}

	// Configure connection pool - smaller for tenant DBs
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping tenant database %s: %w", dbName, err)
	}

	tenantDB := &DB{db}

	// Cache the connection
	tenantMu.Lock()
	tenantConnections[dbName] = tenantDB
	tenantMu.Unlock()

	log.Printf("✅ Connected to tenant database: %s", dbName)
	return tenantDB, nil
}

// GetTenantDB returns cached tenant connection or creates new one
func GetTenantDB(cfg *config.DatabaseConfig, dbName string) (*DB, error) {
	tenantMu.RLock()
	if db, exists := tenantConnections[dbName]; exists {
		tenantMu.RUnlock()
		return db, nil
	}
	tenantMu.RUnlock()

	return ConnectTenant(cfg, dbName)
}

// ConnectMaster connects to the master database for SSA operations
func ConnectMaster(cfg *config.DatabaseConfig, masterDBName string) (*DB, error) {
	masterCfg := &config.DatabaseConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		DBName:   masterDBName,
		SSLMode:  cfg.SSLMode,
	}

	return Connect(masterCfg)
}

// CloseTenantConnections closes all tenant database connections
func CloseTenantConnections() {
	tenantMu.Lock()
	defer tenantMu.Unlock()

	for name, db := range tenantConnections {
		if err := db.Close(); err != nil {
			log.Printf("Error closing tenant DB %s: %v", name, err)
		}
	}
	tenantConnections = make(map[string]*DB)
	log.Println("All tenant connections closed")
}
