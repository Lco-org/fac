package services

import (
	"context"
	"cso-erp/internal/database"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// BackupService handles database backup and restore operations
type BackupService struct {
	db         *database.DB
	backupDir  string
	pgDumpPath string
}

// NewBackupService creates a new backup service
func NewBackupService(db *database.DB, backupDir string) *BackupService {
	// Default pg_dump path for Windows (assumes it's in PATH or standard location)
	pgDump := "pg_dump.exe"

	return &BackupService{
		db:         db,
		backupDir:  backupDir,
		pgDumpPath: pgDump,
	}
}

// BackupRecord represents a backup record
type BackupRecord struct {
	ID        uuid.UUID `json:"id"`
	FileName  string    `json:"file_name"`
	FilePath  string    `json:"file_path"`
	FileSize  int64     `json:"file_size"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateBackup creates a new database backup
func (s *BackupService) CreateBackup(ctx context.Context, dbName string, createdBy *uuid.UUID) (*BackupRecord, error) {
	// Generate filename
	fileName := fmt.Sprintf("backup_%s_%s.sql", dbName, time.Now().Format("20060102_150405"))
	fullPath := filepath.Join(s.backupDir, fileName)

	// Execute pg_dump
	// Note: PGPASSWORD environment variable should be set for authentication
	cmd := exec.CommandContext(ctx, s.pgDumpPath,
		"-h", "localhost",
		"-U", "postgres", // Should assume standard user or configured user
		"-f", fullPath,
		dbName,
	)

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}

	// Create record
	record := &BackupRecord{
		ID:        uuid.New(),
		FileName:  fileName,
		FilePath:  fullPath,
		Status:    "success",
		CreatedAt: time.Now(),
	}

	// Save to DB
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO backup_history (id, file_name, file_path, file_size, status, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, record.ID, record.FileName, record.FilePath, 0, record.Status, createdBy, record.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save backup record: %w", err)
	}

	return record, nil
}

// ListBackups retrieves backup history
func (s *BackupService) ListBackups(ctx context.Context) ([]BackupRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, file_name, file_path, status, created_at
		FROM backup_history
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var backups []BackupRecord
	for rows.Next() {
		var b BackupRecord
		if err := rows.Scan(&b.ID, &b.FileName, &b.FilePath, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, nil
}
