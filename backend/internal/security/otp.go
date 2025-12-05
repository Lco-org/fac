package security

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrOTPExpired indicates the OTP code has expired
	ErrOTPExpired = errors.New("verification code has expired")
	// ErrOTPInvalid indicates the OTP code is incorrect
	ErrOTPInvalid = errors.New("invalid verification code")
	// ErrOTPNotFound indicates no pending OTP exists for this user
	ErrOTPNotFound = errors.New("no pending verification code found")
	// ErrTooManyAttempts indicates too many failed verification attempts
	ErrTooManyAttempts = errors.New("too many failed attempts, please request a new code")
)

const (
	// OTPLength is the number of digits in the OTP code
	OTPLength = 6
	// OTPExpiry is how long the OTP is valid
	OTPExpiry = 10 * time.Minute
	// MaxOTPAttempts is the maximum number of verification attempts
	MaxOTPAttempts = 5
)

// OTPRecord represents a stored OTP verification code
type OTPRecord struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Code      string // Hashed OTP code
	ExpiresAt time.Time
	Attempts  int
	CreatedAt time.Time
}

// OTPService handles One-Time Password generation and verification
type OTPService struct {
	db *sql.DB
}

// NewOTPService creates a new OTP service
func NewOTPService(db *sql.DB) *OTPService {
	return &OTPService{db: db}
}

// GenerateOTP creates a new 6-digit OTP code for the user
// Returns the plain code (for sending via email) and stores a hashed version
func (s *OTPService) GenerateOTP(userID uuid.UUID) (string, error) {
	// Generate secure random 6-digit code
	code, err := generateSecureCode(OTPLength)
	if err != nil {
		return "", err
	}

	// Hash the code before storing (using Argon2id)
	hashedCode, err := HashPassword(code, &Argon2Params{
		Memory:      32 * 1024, // 32MB (lighter for OTP)
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	})
	if err != nil {
		return "", err
	}

	// Delete any existing OTP for this user
	_, _ = s.db.Exec(`DELETE FROM otp_codes WHERE user_id = $1`, userID)

	// Store the new OTP
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

// VerifyOTP checks if the provided code matches the stored OTP
func (s *OTPService) VerifyOTP(userID uuid.UUID, code string) error {
	var record struct {
		ID        uuid.UUID
		CodeHash  string
		ExpiresAt time.Time
		Attempts  int
	}

	err := s.db.QueryRow(`
		SELECT id, code_hash, expires_at, attempts
		FROM otp_codes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(&record.ID, &record.CodeHash, &record.ExpiresAt, &record.Attempts)

	if err == sql.ErrNoRows {
		return ErrOTPNotFound
	}
	if err != nil {
		return err
	}

	// Check attempts
	if record.Attempts >= MaxOTPAttempts {
		// Delete the OTP after too many attempts
		s.db.Exec(`DELETE FROM otp_codes WHERE id = $1`, record.ID)
		return ErrTooManyAttempts
	}

	// Increment attempts
	s.db.Exec(`UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`, record.ID)

	// Check expiry
	if time.Now().After(record.ExpiresAt) {
		s.db.Exec(`DELETE FROM otp_codes WHERE id = $1`, record.ID)
		return ErrOTPExpired
	}

	// Verify code
	valid, err := VerifyPassword(code, record.CodeHash)
	if err != nil || !valid {
		return ErrOTPInvalid
	}

	// Delete OTP after successful verification
	s.db.Exec(`DELETE FROM otp_codes WHERE id = $1`, record.ID)

	return nil
}

// CleanupExpiredOTPs removes all expired OTP codes
func (s *OTPService) CleanupExpiredOTPs() error {
	_, err := s.db.Exec(`DELETE FROM otp_codes WHERE expires_at < NOW()`)
	return err
}

// generateSecureCode generates a cryptographically secure numeric code
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

// ConstantTimeCompare compares two strings in constant time
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
