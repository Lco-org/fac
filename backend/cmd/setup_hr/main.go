package main

import (
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"log"
)

func main() {
	cfg := config.Load()

	// Force connection to the correct DB if needed, but config.Load() should get it.
	// We rely on the internal config logic.

	db, err := database.Connect(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Connected to database. Creating HR tables...")

	query := `
	CREATE TABLE IF NOT EXISTS salary_bonuses (
		id UUID PRIMARY KEY,
		employee_id UUID NOT NULL REFERENCES employees(id),
		month_year TIMESTAMP NOT NULL,
		amount DECIMAL(15, 2) NOT NULL,
		currency VARCHAR(3) NOT NULL,
		reason TEXT NOT NULL,
		reason_ar TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		created_by UUID REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS salary_deductions (
		id UUID PRIMARY KEY,
		employee_id UUID NOT NULL REFERENCES employees(id),
		month_year TIMESTAMP NOT NULL,
		amount DECIMAL(15, 2) NOT NULL,
		currency VARCHAR(3) NOT NULL,
		reason TEXT NOT NULL,
		reason_ar TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		created_by UUID REFERENCES users(id)
	);
	`

	_, err = db.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	log.Println("Successfully created HR tables!")
}
