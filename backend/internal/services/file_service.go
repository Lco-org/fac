package services

import (
	"context"
	"cso-erp/internal/database"
	apperrors "cso-erp/internal/errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// FileService handles file uploads and attachments
type FileService struct {
	db        *database.DB
	uploadDir string
}

// NewFileService creates a new file service
func NewFileService(db *database.DB, uploadDir string) (*FileService, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	return &FileService{
		db:        db,
		uploadDir: uploadDir,
	}, nil
}

// Attachment represents a file attachment
type Attachment struct {
	ID         uuid.UUID
	EntityType string
	EntityID   uuid.UUID
	FileName   string
	FilePath   string
	FileSize   int64
	MimeType   string
	UploadedBy *uuid.UUID
	CreatedAt  time.Time
}

// UploadAttachment handles file upload and database record creation
// tenantSlug ensures each organization has isolated storage
func (s *FileService) UploadAttachment(ctx context.Context, file *multipart.FileHeader, entityType string, entityID uuid.UUID, uploadedBy *uuid.UUID, tenantSlug string) (*Attachment, error) {
	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return nil, apperrors.New("FILE_001", "File size exceeds 10MB limit", 400)
	}

	// Validate tenantSlug to prevent directory traversal
	if tenantSlug == "" || tenantSlug == ".." || filepath.IsAbs(tenantSlug) {
		return nil, apperrors.New("FILE_003", "Invalid tenant identifier", 400)
	}

	// Generate unique filename with tenant isolation
	// Structure: uploads/{tenant_slug}/{entity_type}/{unique_filename}
	ext := filepath.Ext(file.Filename)
	newFileName := fmt.Sprintf("%s_%s%s", entityType, uuid.New().String(), ext)
	relativePath := filepath.Join(tenantSlug, entityType, newFileName)
	fullPath := filepath.Join(s.uploadDir, relativePath)

	// Ensure tenant-specific sub-directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, apperrors.ErrInternal.WithError(err)
	}

	// Save file to disk
	src, err := file.Open()
	if err != nil {
		return nil, apperrors.ErrInternal.WithError(err)
	}
	defer src.Close()

	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, apperrors.ErrInternal.WithError(err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, apperrors.ErrInternal.WithError(err)
	}

	// Create database record
	attachment := &Attachment{
		ID:         uuid.New(),
		EntityType: entityType,
		EntityID:   entityID,
		FileName:   file.Filename,
		FilePath:   relativePath,
		FileSize:   file.Size,
		MimeType:   file.Header.Get("Content-Type"),
		UploadedBy: uploadedBy,
		CreatedAt:  time.Now(),
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO attachments (id, entity_type, entity_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, attachment.ID, attachment.EntityType, attachment.EntityID, attachment.FileName, attachment.FilePath, attachment.FileSize, attachment.MimeType, attachment.UploadedBy, attachment.CreatedAt)

	if err != nil {
		// Cleanup file if DB insert fails
		os.Remove(fullPath)
		return nil, apperrors.ErrDatabaseError.WithError(err)
	}

	return attachment, nil
}

// GetAttachments retrieves all attachments for an entity
func (s *FileService) GetAttachments(ctx context.Context, entityType string, entityID uuid.UUID) ([]Attachment, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, entity_type, entity_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		FROM attachments
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
	`, entityType, entityID)
	if err != nil {
		return nil, apperrors.ErrDatabaseError.WithError(err)
	}
	defer rows.Close()

	var attachments []Attachment
	for rows.Next() {
		var a Attachment
		if err := rows.Scan(&a.ID, &a.EntityType, &a.EntityID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.CreatedAt); err != nil {
			return nil, apperrors.ErrDatabaseError.WithError(err)
		}
		attachments = append(attachments, a)
	}
	return attachments, nil
}

// DeleteAttachment deletes an attachment
func (s *FileService) DeleteAttachment(ctx context.Context, attachmentID uuid.UUID) error {
	// Get file path first
	var filePath string
	err := s.db.QueryRowContext(ctx, `SELECT file_path FROM attachments WHERE id = $1`, attachmentID).Scan(&filePath)
	if err != nil {
		return apperrors.ErrNotFound
	}

	// Delete record from DB
	_, err = s.db.ExecContext(ctx, `DELETE FROM attachments WHERE id = $1`, attachmentID)
	if err != nil {
		return apperrors.ErrDatabaseError.WithError(err)
	}

	// Delete file from disk
	fullPath := filepath.Join(s.uploadDir, filePath)
	os.Remove(fullPath) // Ignore error if file not found

	return nil
}

// GetAttachmentByID retrieves a single attachment
func (s *FileService) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*Attachment, error) {
	var a Attachment
	err := s.db.QueryRowContext(ctx, `
		SELECT id, entity_type, entity_id, file_name, file_path, file_size, mime_type, uploaded_by, created_at
		FROM attachments
		WHERE id = $1
	`, attachmentID).Scan(&a.ID, &a.EntityType, &a.EntityID, &a.FileName, &a.FilePath, &a.FileSize, &a.MimeType, &a.UploadedBy, &a.CreatedAt)

	if err != nil {
		return nil, apperrors.ErrNotFound
	}
	return &a, nil
}
