package security

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrChainBroken indicates audit chain integrity has been compromised
	ErrChainBroken = errors.New("audit chain integrity compromised: hash mismatch detected")
	// ErrLogNotFound indicates the requested audit log was not found
	ErrLogNotFound = errors.New("audit log not found")
	// ErrChainCorrupted indicates missing entries in the audit chain
	ErrChainCorrupted = errors.New("audit chain corrupted: missing entries detected")
)

// AuditAction represents the type of audited action
type AuditAction string

const (
	// Authentication actions
	ActionLogin             AuditAction = "LOGIN"
	ActionLogout            AuditAction = "LOGOUT"
	ActionLoginFailed       AuditAction = "LOGIN_FAILED"
	ActionPasswordChange    AuditAction = "PASSWORD_CHANGE"
	ActionRevokeAllSessions AuditAction = "REVOKE_ALL_SESSIONS"

	// Transaction actions
	ActionTransactionCreate  AuditAction = "TRANSACTION_CREATE"
	ActionTransactionReverse AuditAction = "TRANSACTION_REVERSE"
	ActionTransferCreate     AuditAction = "TRANSFER_CREATE"

	// User management actions
	ActionUserCreate AuditAction = "USER_CREATE"
	ActionUserUpdate AuditAction = "USER_UPDATE"
	ActionUserDelete AuditAction = "USER_DELETE"

	// Data modification actions
	ActionDataModified AuditAction = "DATA_MODIFIED"
	ActionDataDeleted  AuditAction = "DATA_DELETED"
	ActionDataAccess   AuditAction = "DATA_ACCESS"

	// Container actions
	ActionContainerCreate AuditAction = "CONTAINER_CREATE"
	ActionContainerUpdate AuditAction = "CONTAINER_UPDATE"
	ActionContainerDelete AuditAction = "CONTAINER_DELETE"

	// Project actions
	ActionProjectCreate AuditAction = "PROJECT_CREATE"
	ActionProjectUpdate AuditAction = "PROJECT_UPDATE"
	ActionProjectDelete AuditAction = "PROJECT_DELETE"
)

// AuditLog represents a tamper-proof audit log entry
// Each entry contains a hash of the previous entry, creating an immutable chain
type AuditLog struct {
	ID           uuid.UUID   `json:"id"`
	Sequence     int64       `json:"sequence"`
	Action       AuditAction `json:"action"`
	UserID       *uuid.UUID  `json:"user_id,omitempty"`
	Username     *string     `json:"username,omitempty"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	Details      string      `json:"details"` // JSON serialized details
	IPAddress    string      `json:"ip_address"`
	UserAgent    string      `json:"user_agent"`
	PreviousHash string      `json:"previous_hash"`
	CurrentHash  string      `json:"current_hash"`
	CreatedAt    time.Time   `json:"created_at"`
}

// AuditService provides tamper-proof audit logging with blockchain-style chaining
type AuditService struct {
	db *sql.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

// LogAction creates a new chained audit log entry
// Each log includes a hash of the previous log, making tampering detectable
func (s *AuditService) LogAction(ctx context.Context, entry *AuditLog) error {
	if entry == nil {
		return errors.New("audit entry cannot be nil")
	}

	// Get the last log's hash for chaining
	var previousHash string
	var lastSequence int64

	err := s.db.QueryRowContext(ctx, `
		SELECT sequence, current_hash 
		FROM audit_logs 
		ORDER BY sequence DESC 
		LIMIT 1
	`).Scan(&lastSequence, &previousHash)

	if err != nil && err != sql.ErrNoRows {
		return err
	}

	// For genesis block, use special marker
	if err == sql.ErrNoRows {
		previousHash = "GENESIS"
		lastSequence = 0
	}

	entry.ID = uuid.New()
	entry.Sequence = lastSequence + 1
	entry.PreviousHash = previousHash
	entry.CreatedAt = time.Now().UTC()

	// Calculate current hash: SHA-256(sequence + action + details + previous_hash + timestamp)
	entry.CurrentHash = s.calculateHash(entry)

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO audit_logs (
			id, sequence, action, user_id, username, resource_type, 
			resource_id, details, ip_address, user_agent, 
			previous_hash, current_hash, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		entry.ID,
		entry.Sequence,
		entry.Action,
		entry.UserID,
		entry.Username,
		entry.ResourceType,
		entry.ResourceID,
		entry.Details,
		entry.IPAddress,
		entry.UserAgent,
		entry.PreviousHash,
		entry.CurrentHash,
		entry.CreatedAt,
	)

	return err
}

// calculateHash generates SHA-256 hash for the audit entry
// This creates a unique fingerprint that changes if any data is modified
func (s *AuditService) calculateHash(entry *AuditLog) string {
	data := struct {
		Sequence     int64       `json:"seq"`
		Action       AuditAction `json:"act"`
		UserID       *uuid.UUID  `json:"uid"`
		ResourceID   *uuid.UUID  `json:"rid"`
		ResourceType string      `json:"rtype"`
		Details      string      `json:"det"`
		PreviousHash string      `json:"prev"`
		Timestamp    string      `json:"ts"`
	}{
		Sequence:     entry.Sequence,
		Action:       entry.Action,
		UserID:       entry.UserID,
		ResourceID:   entry.ResourceID,
		ResourceType: entry.ResourceType,
		Details:      entry.Details,
		PreviousHash: entry.PreviousHash,
		Timestamp:    entry.CreatedAt.Format(time.RFC3339Nano),
	}

	jsonData, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:])
}

// VerifyChainIntegrity validates the entire audit chain
// Returns true if all entries are valid and properly chained
// Any modification, deletion, or insertion will be detected
func (s *AuditService) VerifyChainIntegrity(ctx context.Context) (bool, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, sequence, action, user_id, username, resource_type,
		       resource_id, details, ip_address, user_agent,
		       previous_hash, current_hash, created_at
		FROM audit_logs
		ORDER BY sequence ASC
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var previousHash = "GENESIS"
	var expectedSequence int64 = 1

	for rows.Next() {
		var entry AuditLog
		err := rows.Scan(
			&entry.ID,
			&entry.Sequence,
			&entry.Action,
			&entry.UserID,
			&entry.Username,
			&entry.ResourceType,
			&entry.ResourceID,
			&entry.Details,
			&entry.IPAddress,
			&entry.UserAgent,
			&entry.PreviousHash,
			&entry.CurrentHash,
			&entry.CreatedAt,
		)
		if err != nil {
			return false, err
		}

		// Verify sequence continuity (detect deletions)
		if entry.Sequence != expectedSequence {
			return false, ErrChainCorrupted
		}

		// Verify chain linkage
		if entry.PreviousHash != previousHash {
			return false, ErrChainBroken
		}

		// Verify hash integrity (detect modifications)
		calculatedHash := s.calculateHash(&entry)
		if calculatedHash != entry.CurrentHash {
			return false, ErrChainBroken
		}

		previousHash = entry.CurrentHash
		expectedSequence++
	}

	if err := rows.Err(); err != nil {
		return false, err
	}

	return true, nil
}

// GetAuditLogs retrieves audit logs with optional filtering
func (s *AuditService) GetAuditLogs(ctx context.Context, userID *uuid.UUID, action *AuditAction, limit, offset int) ([]AuditLog, error) {
	query := `
		SELECT id, sequence, action, user_id, username, resource_type,
		       resource_id, details, ip_address, user_agent,
		       previous_hash, current_hash, created_at
		FROM audit_logs
		WHERE ($1::uuid IS NULL OR user_id = $1)
		  AND ($2::text IS NULL OR action = $2)
		ORDER BY sequence DESC
		LIMIT $3 OFFSET $4
	`

	var actionStr *string
	if action != nil {
		s := string(*action)
		actionStr = &s
	}

	rows, err := s.db.QueryContext(ctx, query, userID, actionStr, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var entry AuditLog
		err := rows.Scan(
			&entry.ID,
			&entry.Sequence,
			&entry.Action,
			&entry.UserID,
			&entry.Username,
			&entry.ResourceType,
			&entry.ResourceID,
			&entry.Details,
			&entry.IPAddress,
			&entry.UserAgent,
			&entry.PreviousHash,
			&entry.CurrentHash,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, entry)
	}

	return logs, rows.Err()
}

// AuditDetails helper to create JSON details from a map
func AuditDetails(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}

// CreateAuditEntry is a helper to create a basic audit entry
func CreateAuditEntry(action AuditAction, resourceType string, userID *uuid.UUID, username *string, ipAddress, userAgent string) *AuditLog {
	return &AuditLog{
		Action:       action,
		UserID:       userID,
		Username:     username,
		ResourceType: resourceType,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      "{}",
	}
}
