package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/logger"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ActivityService handles recording of user activities
type ActivityService struct {
	db *database.DB
}

// NewActivityService creates a new activity service
func NewActivityService(db *database.DB) *ActivityService {
	return &ActivityService{db: db}
}

// LogActivity saves an activity record
func (s *ActivityService) LogActivity(ctx context.Context, action, entityType string, entityID uuid.UUID, userID *uuid.UUID, username string, details map[string]interface{}, ip string) {
	// Execute asynchronously to not block the main request
	go func() {
		// Create a new context with timeout for the background operation
		bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var detailsJSON []byte
		var err error
		if details != nil {
			detailsJSON, err = json.Marshal(details)
			if err != nil {
				logger.Error("Failed to marshal activity details: %v", err)
				return
			}
		}

		_, err = s.db.ExecContext(bgCtx, `
			INSERT INTO activity_logs (id, action, entity_type, entity_id, user_id, username, details, ip, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, uuid.New(), action, entityType, entityID, userID, username, detailsJSON, ip, time.Now())

		if err != nil {
			logger.Error("Failed to log activity: %v", err)
		}
	}()
}

// ActivityLog represents a log entry
type ActivityLog struct {
	ID         uuid.UUID              `json:"id"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   uuid.UUID              `json:"entity_id"`
	UserID     *uuid.UUID             `json:"user_id"`
	Username   string                 `json:"username"`
	Details    map[string]interface{} `json:"details"`
	IP         string                 `json:"ip"`
	CreatedAt  time.Time              `json:"created_at"`
}

// GetEntityActivities retrieves activities for a specific entity
func (s *ActivityService) GetEntityActivities(ctx context.Context, entityType string, entityID uuid.UUID) ([]ActivityLog, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, action, entity_type, entity_id, user_id, username, details, ip, created_at
		FROM activity_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT 100
	`, entityType, entityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ActivityLog
	for rows.Next() {
		var l ActivityLog
		var detailsJSON []byte
		if err := rows.Scan(&l.ID, &l.Action, &l.EntityType, &l.EntityID, &l.UserID, &l.Username, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		if len(detailsJSON) > 0 {
			json.Unmarshal(detailsJSON, &l.Details)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

// GetUserActivities retrieves activities for a specific user
func (s *ActivityService) GetUserActivities(ctx context.Context, userID uuid.UUID) ([]ActivityLog, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, action, entity_type, entity_id, user_id, username, details, ip, created_at
		FROM activity_logs
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ActivityLog
	for rows.Next() {
		var l ActivityLog
		var detailsJSON []byte
		if err := rows.Scan(&l.ID, &l.Action, &l.EntityType, &l.EntityID, &l.UserID, &l.Username, &detailsJSON, &l.IP, &l.CreatedAt); err != nil {
			return nil, err
		}
		if len(detailsJSON) > 0 {
			json.Unmarshal(detailsJSON, &l.Details)
		}
		logs = append(logs, l)
	}
	return logs, nil
}
