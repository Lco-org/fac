package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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
		return db.DB.Close()
	}
	return nil
}
