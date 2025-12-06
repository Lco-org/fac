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
	ErrChainBroken    = errors.New("audit chain integrity compromised: hash mismatch detected")
	ErrLogNotFound    = errors.New("audit log not found")
	ErrChainCorrupted = errors.New("audit chain corrupted: missing entries detected")
)

type AuditAction string

const (
	ActionLogin             AuditAction = "LOGIN"
	ActionLogout            AuditAction = "LOGOUT"
	ActionLoginFailed       AuditAction = "LOGIN_FAILED"
	ActionPasswordChange    AuditAction = "PASSWORD_CHANGE"
	ActionRevokeAllSessions AuditAction = "REVOKE_ALL_SESSIONS"
	ActionUserCreate        AuditAction = "USER_CREATE"
	ActionDataModified      AuditAction = "DATA_MODIFIED"
)

type AuditLog struct {
	ID           uuid.UUID   `json:"id"`
	Sequence     int64       `json:"sequence"`
	Action       AuditAction `json:"action"`
	UserID       *uuid.UUID  `json:"user_id,omitempty"`
	Username     *string     `json:"username,omitempty"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	Details      string      `json:"details"`
	IPAddress    string      `json:"ip_address"`
	UserAgent    string      `json:"user_agent"`
	PreviousHash string      `json:"previous_hash"`
	CurrentHash  string      `json:"current_hash"`
	CreatedAt    time.Time   `json:"created_at"`
}

type AuditService struct {
	db *sql.DB
}

func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) LogAction(ctx context.Context, entry *AuditLog) error {
	if entry == nil {
		return errors.New("audit entry cannot be nil")
	}

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

	if err == sql.ErrNoRows {
		previousHash = "GENESIS"
		lastSequence = 0
	}

	entry.ID = uuid.New()
	entry.Sequence = lastSequence + 1
	entry.PreviousHash = previousHash
	entry.CreatedAt = time.Now().UTC()

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

		if entry.Sequence != expectedSequence {
			return false, ErrChainCorrupted
		}

		if entry.PreviousHash != previousHash {
			return false, ErrChainBroken
		}

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

func AuditDetails(data map[string]interface{}) string {
	if data == nil {
		return "{}"
	}
	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}
