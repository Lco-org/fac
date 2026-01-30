package services

import (
	"context"
	"cso-erp/internal/database"
	"cso-erp/internal/logger"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NotificationService handles notifications and expiry alerts
type NotificationService struct {
	db           *database.DB
	emailService *EmailService
	stopChan     chan struct{}
	wg           sync.WaitGroup
	isRunning    bool
	mu           sync.Mutex
}

// NewNotificationService creates a new notification service
func NewNotificationService(db *database.DB, emailService *EmailService) *NotificationService {
	return &NotificationService{
		db:           db,
		emailService: emailService,
		stopChan:     make(chan struct{}),
	}
}

// Notification represents a notification record
type Notification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// ExpiringMember represents a member with expiring membership
type ExpiringMember struct {
	ID           uuid.UUID
	Name         string
	Email        string
	MemberType   string
	ExpiryDate   time.Time
	DaysUntil    int
	ContactEmail string
}

// StartExpiryChecker starts the background worker to check for expiring memberships
func (s *NotificationService) StartExpiryChecker(checkInterval time.Duration) {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// Run immediately on start
		s.checkExpiringMemberships()

		ticker := time.NewTicker(checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.checkExpiringMemberships()
			case <-s.stopChan:
				logger.Info("Expiry checker stopped")
				return
			}
		}
	}()

	logger.Info("Expiry checker started with interval: %v", checkInterval)
}

// StopExpiryChecker stops the background worker
func (s *NotificationService) StopExpiryChecker() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
}

// checkExpiringMemberships checks for members with expiring memberships
func (s *NotificationService) checkExpiringMemberships() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find members expiring in the next 30 days
	expiringMembers, err := s.getExpiringMembers(ctx, 30)
	if err != nil {
		logger.Error("Failed to get expiring members: %v", err)
		return
	}

	logger.Info("Found %d members with expiring memberships", len(expiringMembers))

	for _, member := range expiringMembers {
		// Create notification record
		if err := s.createNotification(ctx, member); err != nil {
			logger.Error("Failed to create notification for member %s: %v", member.ID, err)
			continue
		}

		// Send email notification
		if s.emailService != nil && member.Email != "" {
			if err := s.sendExpiryEmail(ctx, member); err != nil {
				logger.Error("Failed to send expiry email to %s: %v", member.Email, err)
			}
		}
	}
}

// getExpiringMembers retrieves members with expiring memberships
func (s *NotificationService) getExpiringMembers(ctx context.Context, daysAhead int) ([]ExpiringMember, error) {
	query := `
		SELECT 
			m.id, 
			m.full_name as name, 
			COALESCE(m.email, '') as email,
			m.member_type,
			m.valid_until,
			EXTRACT(DAY FROM m.valid_until - CURRENT_DATE)::int as days_until
		FROM members m
		WHERE m.is_active = true
			AND m.valid_until IS NOT NULL
			AND m.valid_until BETWEEN CURRENT_DATE AND CURRENT_DATE + $1 * INTERVAL '1 day'
			AND m.id NOT IN (
				SELECT DISTINCT (details->>'member_id')::uuid
				FROM notifications
				WHERE type = 'expiry_warning'
					AND created_at > CURRENT_DATE - INTERVAL '7 days'
					AND details->>'member_id' IS NOT NULL
			)
		ORDER BY m.valid_until ASC
	`

	rows, err := s.db.QueryContext(ctx, query, daysAhead)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ExpiringMember
	for rows.Next() {
		var m ExpiringMember
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.MemberType, &m.ExpiryDate, &m.DaysUntil); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, nil
}

// createNotification creates a notification record
func (s *NotificationService) createNotification(ctx context.Context, member ExpiringMember) error {
	title := fmt.Sprintf("تنبيه: انتهاء عضوية %s", member.Name)
	message := fmt.Sprintf("ستنتهي عضوية %s خلال %d يوم", member.Name, member.DaysUntil)
	details := fmt.Sprintf(`{"member_id": "%s"}`, member.ID.String())

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO notifications (id, user_id, type, title, message, details, is_read, created_at)
		SELECT $1, u.id, $2, $3, $4, $5::jsonb, false, $6
		FROM users u 
		WHERE u.role = 'admin'
		LIMIT 1
	`, uuid.New(), "expiry_warning", title, message, details, time.Now())

	return err
}

// sendExpiryEmail sends an expiry warning email
func (s *NotificationService) sendExpiryEmail(ctx context.Context, member ExpiringMember) error {
	subject := "تذكير: اقتراب انتهاء العضوية"
	body := fmt.Sprintf(`
		<html>
		<body style="font-family: Arial, sans-serif; direction: rtl;">
			<h2>عزيزي %s،</h2>
			<p>نود إعلامك بأن عضويتك ستنتهي خلال <strong>%d يوم</strong>.</p>
			<p>تاريخ انتهاء العضوية: <strong>%s</strong></p>
			<p>يرجى التواصل مع الإدارة لتجديد عضويتك.</p>
			<hr>
			<p style="color: #888; font-size: 12px;">CSO-ERP System</p>
		</body>
		</html>
	`, member.Name, member.DaysUntil, member.ExpiryDate.Format("2006-01-02"))

	return s.emailService.SendEmail(ctx, member.Email, subject, body, true)
}

// ============================================================================
// NOTIFICATION CRUD OPERATIONS
// ============================================================================

// GetUserNotifications retrieves notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID uuid.UUID) ([]Notification, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, user_id, type, title, message, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []Notification
	for rows.Next() {
		var n Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Message, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, n)
	}
	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *NotificationService) MarkAsRead(ctx context.Context, notificationID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE notifications SET is_read = true WHERE id = $1
	`, notificationID)
	return err
}

// MarkAllAsRead marks all user notifications as read
func (s *NotificationService) MarkAllAsRead(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE notifications SET is_read = true WHERE user_id = $1
	`, userID)
	return err
}

// GetUnreadCount returns count of unread notifications
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM notifications WHERE user_id = $1 AND is_read = false
	`, userID).Scan(&count)
	return count, err
}
