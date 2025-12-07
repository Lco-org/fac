package services

import (
	"database/sql"

	"cso-erp/internal/security"
)

// OTPCleanupService wraps the security.OTPService to expose cleanup functionality
// This is used by main.go to start the background cleanup scheduler
type OTPCleanupService struct {
	otpService *security.OTPService
}

// NewOTPCleanupService creates a new OTP cleanup service
func NewOTPCleanupService(db *sql.DB) *OTPCleanupService {
	return &OTPCleanupService{
		otpService: security.NewOTPService(db),
	}
}

// StartCleanupScheduler starts the background OTP cleanup goroutine
func (s *OTPCleanupService) StartCleanupScheduler() {
	s.otpService.StartCleanupScheduler()
}

// StopCleanupScheduler gracefully stops the cleanup scheduler
func (s *OTPCleanupService) StopCleanupScheduler() {
	s.otpService.StopCleanupScheduler()
}
