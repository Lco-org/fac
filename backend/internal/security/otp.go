package security

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"log"
	"math/big"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOTPExpired      = errors.New("verification code has expired")
	ErrOTPInvalid      = errors.New("invalid verification code")
	ErrOTPNotFound     = errors.New("no pending verification code found")
	ErrTooManyAttempts = errors.New("too many failed attempts, please request a new code")
	ErrOTPRateLimited  = errors.New("too many OTP requests, please wait before requesting another code")
)

const (
	OTPLength          = 6
	OTPExpiry          = 10 * time.Minute
	MaxOTPAttempts     = 5
	MaxOTPRequestsHour = 5               // Maximum OTP requests per hour per user
	CleanupInterval    = 5 * time.Minute // Background cleanup interval
)

type OTPService struct {
	db          *sql.DB
	stopCleanup chan struct{}
	cleanupOnce sync.Once
	cleanupWg   sync.WaitGroup
}

func NewOTPService(db *sql.DB) *OTPService {
	return &OTPService{
		db:          db,
		stopCleanup: make(chan struct{}),
	}
}

func (s *OTPService) GenerateOTP(userID uuid.UUID) (string, error) {
	// Check rate limiting - count requests in the last hour
	var recentCount int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM otp_codes 
		WHERE user_id = $1 AND created_at > NOW() - INTERVAL '1 hour'
	`, userID).Scan(&recentCount)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if recentCount >= MaxOTPRequestsHour {
		return "", ErrOTPRateLimited
	}

	// Generate secure code
	code, err := generateSecureCode(OTPLength)
	if err != nil {
		return "", err
	}

	// Hash the OTP code with lighter params (for quick verification)
	hashedCode, err := HashPassword(code, &Argon2Params{
		Memory:      32 * 1024,
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		return "", err
	}

	// Delete any existing unexpired OTP for this user
	_, _ = s.db.Exec(`DELETE FROM otp_codes WHERE user_id = $1 AND expires_at > NOW()`, userID)

	// Insert new OTP
	expiresAt := time.Now().Add(OTPExpiry)
	_, err = s.db.Exec(`
		INSERT INTO otp_codes (id, user_id, code_hash, expires_at, attempts, created_at)
		VALUES ($1, $2, $3, $4, 0, NOW())
	`, uuid.New(), userID, hashedCode, expiresAt)

	if err != nil {
		return "", err
	}

	return code, nil
}

// VerifyOTP verifies an OTP code within a database transaction for ACID compliance.
// Uses row-level locking (FOR UPDATE) to prevent race conditions.
func (s *OTPService) VerifyOTP(userID uuid.UUID, code string) error {
	ctx := context.Background()

	// Start transaction for ACID compliance
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var record struct {
		ID        uuid.UUID
		CodeHash  string
		ExpiresAt time.Time
		Attempts  int
	}

	// Use FOR UPDATE to lock the row during verification (prevent race conditions)
	err = tx.QueryRowContext(ctx, `
		SELECT id, code_hash, expires_at, attempts
		FROM otp_codes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
		FOR UPDATE
	`, userID).Scan(&record.ID, &record.CodeHash, &record.ExpiresAt, &record.Attempts)

	if err == sql.ErrNoRows {
		return ErrOTPNotFound
	}
	if err != nil {
		return err
	}

	// Check expiry FIRST (before counting as an attempt)
	if time.Now().After(record.ExpiresAt) {
		tx.ExecContext(ctx, `DELETE FROM otp_codes WHERE id = $1`, record.ID)
		tx.Commit() // Commit the deletion
		return ErrOTPExpired
	}

	// Check if too many attempts already
	if record.Attempts >= MaxOTPAttempts {
		tx.ExecContext(ctx, `DELETE FROM otp_codes WHERE id = $1`, record.ID)
		tx.Commit() // Commit the deletion
		return ErrTooManyAttempts
	}

	// Verify the code
	valid, err := VerifyPassword(code, record.CodeHash)
	if err != nil || !valid {
		// Increment attempt count on failed verification
		_, err = tx.ExecContext(ctx, `UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`, record.ID)
		if err != nil {
			return err
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return commitErr
		}
		return ErrOTPInvalid
	}

	// Success - delete the OTP code
	_, err = tx.ExecContext(ctx, `DELETE FROM otp_codes WHERE id = $1`, record.ID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// CleanupExpiredOTPs removes all expired OTP codes from the database.
func (s *OTPService) CleanupExpiredOTPs() error {
	result, err := s.db.Exec(`DELETE FROM otp_codes WHERE expires_at < NOW()`)
	if err != nil {
		return err
	}
	if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
		log.Printf("[OTP Cleanup] Removed %d expired OTP codes", rowsAffected)
	}
	return nil
}

// StartCleanupScheduler starts a background goroutine that periodically cleans up expired OTPs.
// This should be called once during application startup.
// Call StopCleanupScheduler to gracefully stop the cleanup process.
func (s *OTPService) StartCleanupScheduler() {
	s.cleanupOnce.Do(func() {
		s.cleanupWg.Add(1)
		go func() {
			defer s.cleanupWg.Done()
			ticker := time.NewTicker(CleanupInterval)
			defer ticker.Stop()

			log.Printf("[OTP Cleanup] Started background cleanup scheduler (interval: %v)", CleanupInterval)

			// Initial cleanup on startup
			if err := s.CleanupExpiredOTPs(); err != nil {
				log.Printf("[OTP Cleanup] Initial cleanup error: %v", err)
			}

			for {
				select {
				case <-ticker.C:
					if err := s.CleanupExpiredOTPs(); err != nil {
						log.Printf("[OTP Cleanup] Scheduled cleanup error: %v", err)
					}
				case <-s.stopCleanup:
					log.Printf("[OTP Cleanup] Stopping background cleanup scheduler")
					return
				}
			}
		}()
	})
}

// StopCleanupScheduler gracefully stops the background cleanup scheduler.
// This should be called during application shutdown.
func (s *OTPService) StopCleanupScheduler() {
	close(s.stopCleanup)
	s.cleanupWg.Wait()
}

func generateSecureCode(length int) (string, error) {
	const digits = "0123456789"
	code := make([]byte, length)

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}
