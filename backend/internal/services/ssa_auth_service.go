package services

import (
	"context"
	"crypto/rand"
	"cso-erp/internal/config"
	"cso-erp/internal/database"
	"cso-erp/internal/middleware"
	"cso-erp/internal/models"
	"cso-erp/internal/security"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// SSAAuthService handles Super System Admin authentication
type SSAAuthService struct {
	masterDB     *database.DB
	cfg          *config.Config
	emailService *EmailService
}

// NewSSAAuthService creates a new SSA auth service
func NewSSAAuthService(masterDB *database.DB, cfg *config.Config, emailService *EmailService) *SSAAuthService {
	return &SSAAuthService{
		masterDB:     masterDB,
		cfg:          cfg,
		emailService: emailService,
	}
}

// SSALoginRequest for SSA login
type SSALoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// SSALoginResult for login response
type SSALoginResult struct {
	Admin            *models.SuperSystemAdmin `json:"admin,omitempty"`
	Token            string                   `json:"token,omitempty"`
	Requires2FA      bool                     `json:"requires_2fa"`
	PendingSessionID string                   `json:"pending_session_id,omitempty"`
	Message          string                   `json:"message,omitempty"`
}

// SSA2FASession stores pending 2FA verification
type SSA2FASession struct {
	AdminID   uuid.UUID
	Code      string
	ExpiresAt time.Time
	Attempts  int
}

// pending2FASessions stores temporary 2FA sessions (in production use Redis/DB)
var pending2FASessions = make(map[string]*SSA2FASession)

// Login authenticates an SSA user with username and password
func (s *SSAAuthService) Login(ctx context.Context, req *SSALoginRequest, clientIP string) (*SSALoginResult, error) {
	// Get admin by username from master database
	admin, err := s.getAdminByUsername(ctx, req.Username)
	if err != nil {
		log.Printf("SSA Login failed for %s: %v", req.Username, err)
		return nil, fmt.Errorf("اسم المستخدم أو كلمة المرور غير صحيحة")
	}
	if admin == nil {
		log.Printf("SSA Login: user not found: %s", req.Username)
		return nil, fmt.Errorf("اسم المستخدم أو كلمة المرور غير صحيحة")
	}

	// Check if admin is active
	if !admin.IsActive {
		log.Printf("SSA Login: inactive admin: %s", req.Username)
		return nil, fmt.Errorf("الحساب معطل. تواصل مع مسؤول النظام")
	}

	// Verify password using Argon2id
	valid, err := security.VerifyPassword(req.Password, admin.PasswordHash)
	if err != nil || !valid {
		log.Printf("SSA Login: invalid password for: %s", req.Username)
		// Update failed login count
		s.recordFailedLogin(ctx, admin.ID)
		return nil, fmt.Errorf("اسم المستخدم أو كلمة المرور غير صحيحة")
	}

	// Check if 2FA is enabled
	if admin.TwoFactorEnabled {
		// Password correct - now send 2FA code
		sessionID, err := s.initiate2FA(ctx, admin)
		if err != nil {
			log.Printf("SSA Login: failed to send 2FA for %s: %v", req.Username, err)
			return nil, fmt.Errorf("فشل إرسال رمز التحقق. حاول مرة أخرى")
		}

		s.logAuditEvent(ctx, admin.ID, "ssa_login_password_verified", clientIP, "Password verified, 2FA sent")

		return &SSALoginResult{
			Requires2FA:      true,
			PendingSessionID: sessionID,
			Message:          "تم إرسال رمز التحقق إلى بريدك الإلكتروني",
		}, nil
	}

	// 2FA not enabled - login directly
	token, err := middleware.GenerateToken(&models.User{
		ID:           admin.ID,
		Username:     admin.Username,
		Role:         models.RoleSystemAdmin, // Use SystemAdmin role for SSA
		TokenVersion: 1,                      // SSA doesn't track token version yet
	}, &s.cfg.JWT)

	if err != nil {
		return nil, fmt.Errorf("فشل إنشاء رمز الدخول")
	}

	s.logAuditEvent(ctx, admin.ID, "ssa_login_success", clientIP, "Login successful (No 2FA)")

	return &SSALoginResult{
		Admin:       admin,
		Token:       token,
		Requires2FA: false,
	}, nil
}

// initiate2FA generates and sends 2FA code
func (s *SSAAuthService) initiate2FA(ctx context.Context, admin *models.SuperSystemAdmin) (string, error) {
	// Generate 6-digit OTP code
	code := generateOTPCode(6)

	// Generate session ID
	sessionID := uuid.New().String()

	// Store session (expires in 10 minutes)
	pending2FASessions[sessionID] = &SSA2FASession{
		AdminID:   admin.ID,
		Code:      code,
		ExpiresAt: time.Now().Add(10 * time.Minute),
		Attempts:  0,
	}

	// Send 2FA code via email
	err := s.emailService.SendSSA2FACode(ctx, admin.Email, admin.FullName, code)
	if err != nil {
		delete(pending2FASessions, sessionID)
		return "", fmt.Errorf("فشل إرسال رمز التحقق: %w", err)
	}

	log.Printf("SSA 2FA code sent to %s (session: %s)", admin.Email, sessionID[:8])
	return sessionID, nil
}

// Verify2FA verifies the 2FA code and completes login
func (s *SSAAuthService) Verify2FA(ctx context.Context, sessionID, code, clientIP string) (*SSALoginResult, error) {
	// Get pending session
	session, exists := pending2FASessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("جلسة التحقق منتهية أو غير صالحة")
	}

	// Check expiry
	if time.Now().After(session.ExpiresAt) {
		delete(pending2FASessions, sessionID)
		return nil, fmt.Errorf("انتهت صلاحية رمز التحقق")
	}

	// Check attempts
	session.Attempts++
	if session.Attempts > 3 {
		delete(pending2FASessions, sessionID)
		return nil, fmt.Errorf("تجاوزت الحد الأقصى للمحاولات")
	}

	// Verify code
	if session.Code != code {
		return nil, fmt.Errorf("رمز التحقق غير صحيح. المحاولات المتبقية: %d", 3-session.Attempts)
	}

	// Code verified - get admin
	admin, err := s.GetAdminByID(ctx, session.AdminID)
	if err != nil || admin == nil {
		delete(pending2FASessions, sessionID)
		return nil, fmt.Errorf("خطأ في التحقق من الحساب")
	}

	// Generate JWT token
	token, err := s.generateSSAToken(admin)
	if err != nil {
		return nil, fmt.Errorf("فشل إنشاء رمز الدخول")
	}

	// Update last login
	s.updateLastLogin(ctx, admin.ID)

	// Log successful login
	s.logAuditEvent(ctx, admin.ID, "ssa_login_complete", clientIP, "2FA verified, login successful")

	// Clean up session
	delete(pending2FASessions, sessionID)

	// Remove sensitive data
	admin.PasswordHash = ""
	admin.TwoFactorSecret = nil

	return &SSALoginResult{
		Admin:       admin,
		Token:       token,
		Requires2FA: false,
		Message:     "تم تسجيل الدخول بنجاح",
	}, nil
}

// generateSSAToken generates JWT for SSA
func (s *SSAAuthService) generateSSAToken(admin *models.SuperSystemAdmin) (string, error) {
	claims := jwt.MapClaims{
		"sub":      admin.ID.String(),
		"username": admin.Username,
		"email":    admin.Email,
		"role":     "system_admin",
		"type":     "ssa",
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Hour * 4).Unix(), // 4 hour expiry for SSA
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}

// getAdminByUsername gets admin from master database
func (s *SSAAuthService) getAdminByUsername(ctx context.Context, username string) (*models.SuperSystemAdmin, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, 
		       two_factor_enabled, two_factor_secret, last_login, 
		       failed_login_attempts, locked_until, created_at, updated_at
		FROM super_system_admins
		WHERE username = $1
	`

	var admin models.SuperSystemAdmin
	var twoFactorSecret sql.NullString
	var lastLogin, lockedUntil sql.NullTime

	err := s.masterDB.QueryRowContext(ctx, query, username).Scan(
		&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
		&admin.FullName, &admin.IsActive, &admin.TwoFactorEnabled,
		&twoFactorSecret, &lastLogin, &admin.FailedLoginAttempts,
		&lockedUntil, &admin.CreatedAt, &admin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if twoFactorSecret.Valid {
		admin.TwoFactorSecret = &twoFactorSecret.String
	}
	if lastLogin.Valid {
		admin.LastLogin = &lastLogin.Time
	}
	if lockedUntil.Valid {
		admin.LockedUntil = &lockedUntil.Time
	}

	return &admin, nil
}

// GetAdminByID gets admin by ID
func (s *SSAAuthService) GetAdminByID(ctx context.Context, id uuid.UUID) (*models.SuperSystemAdmin, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, is_active, 
		       two_factor_enabled, created_at, updated_at
		FROM super_system_admins
		WHERE id = $1
	`

	var admin models.SuperSystemAdmin
	err := s.masterDB.QueryRowContext(ctx, query, id).Scan(
		&admin.ID, &admin.Username, &admin.Email, &admin.PasswordHash,
		&admin.FullName, &admin.IsActive, &admin.TwoFactorEnabled,
		&admin.CreatedAt, &admin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &admin, nil
}

// recordFailedLogin records a failed login attempt
func (s *SSAAuthService) recordFailedLogin(ctx context.Context, adminID uuid.UUID) {
	query := `
		UPDATE super_system_admins 
		SET failed_login_attempts = failed_login_attempts + 1,
		    locked_until = CASE 
		        WHEN failed_login_attempts >= 4 THEN NOW() + INTERVAL '1 hour'
		        ELSE locked_until
		    END,
		    updated_at = NOW()
		WHERE id = $1
	`
	s.masterDB.ExecContext(ctx, query, adminID)
}

// updateLastLogin updates last login timestamp
func (s *SSAAuthService) updateLastLogin(ctx context.Context, adminID uuid.UUID) {
	query := `
		UPDATE super_system_admins 
		SET last_login = NOW(), 
		    failed_login_attempts = 0,
		    locked_until = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`
	s.masterDB.ExecContext(ctx, query, adminID)
}

// logAuditEvent logs SSA action
func (s *SSAAuthService) logAuditEvent(ctx context.Context, adminID uuid.UUID, action, ipAddress, details string) {
	query := `
		INSERT INTO super_admin_audit (admin_id, admin_email, action, resource_type, ip_address, created_at)
		SELECT $1, email, $2, 'system', $3::inet, NOW()
		FROM super_system_admins WHERE id = $1
	`
	s.masterDB.ExecContext(ctx, query, adminID, action, ipAddress)
}

// SSAAuditLog represents an audit log entry
type SSAAuditLog struct {
	ID           uuid.UUID  `json:"id"`
	AdminID      uuid.UUID  `json:"admin_id"`
	AdminEmail   string     `json:"admin_email"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *uuid.UUID `json:"resource_id,omitempty"`
	IPAddress    string     `json:"ip_address"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GetAuditLogs retrieves SSA audit logs
func (s *SSAAuthService) GetAuditLogs(ctx context.Context, limit, offset int) ([]SSAAuditLog, int, error) {
	// Get total count
	var total int
	err := s.masterDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM super_admin_audit").Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, admin_id, admin_email, action, resource_type, resource_id, 
		       host(ip_address) as ip, created_at
		FROM super_admin_audit
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.masterDB.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []SSAAuditLog
	for rows.Next() {
		var log SSAAuditLog
		var resourceID *string
		var ip *string

		if err := rows.Scan(
			&log.ID, &log.AdminID, &log.AdminEmail, &log.Action,
			&log.ResourceType, &resourceID, &ip, &log.CreatedAt,
		); err != nil {
			return nil, 0, err
		}

		if resourceID != nil {
			uid, _ := uuid.Parse(*resourceID)
			log.ResourceID = &uid
		}
		if ip != nil {
			log.IPAddress = *ip
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

// generateOTPCode generates a random numeric OTP
func generateOTPCode(length int) string {
	const digits = "0123456789"
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		code[i] = digits[n.Int64()]
	}
	return string(code)
}

// maskEmail masks email for display
func maskEmail(email string) string {
	if len(email) < 5 {
		return "***"
	}
	atIndex := -1
	for i, c := range email {
		if c == '@' {
			atIndex = i
			break
		}
	}
	if atIndex < 2 {
		return "***" + email[atIndex:]
	}
	return email[:2] + "***" + email[atIndex:]
}
