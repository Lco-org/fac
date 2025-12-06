package security

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
)

var (
	ErrOTPExpired      = errors.New("verification code has expired")
	ErrOTPInvalid      = errors.New("invalid verification code")
	ErrOTPNotFound     = errors.New("no pending verification code found")
	ErrTooManyAttempts = errors.New("too many failed attempts, please request a new code")
)

const (
	OTPLength      = 6
	OTPExpiry      = 10 * time.Minute
	MaxOTPAttempts = 5
)

type OTPService struct {
	db *sql.DB
}

func NewOTPService(db *sql.DB) *OTPService {
	return &OTPService{db: db}
}

func (s *OTPService) GenerateOTP(userID uuid.UUID) (string, error) {
	code, err := generateSecureCode(OTPLength)
	if err != nil {
		return "", err
	}

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

	_, _ = s.db.Exec(`DELETE FROM otp_codes WHERE user_id = $1`, userID)

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

	if record.Attempts >= MaxOTPAttempts {
		s.db.Exec(`DELETE FROM otp_codes WHERE id = $1`, record.ID)
		return ErrTooManyAttempts
	}

	s.db.Exec(`UPDATE otp_codes SET attempts = attempts + 1 WHERE id = $1`, record.ID)

	if time.Now().After(record.ExpiresAt) {
		s.db.Exec(`DELETE FROM otp_codes WHERE id = $1`, record.ID)
		return ErrOTPExpired
	}

	valid, err := VerifyPassword(code, record.CodeHash)
	if err != nil || !valid {
		return ErrOTPInvalid
	}

	s.db.Exec(`DELETE FROM otp_codes WHERE id = $1`, record.ID)

	return nil
}

func (s *OTPService) CleanupExpiredOTPs() error {
	_, err := s.db.Exec(`DELETE FROM otp_codes WHERE expires_at < NOW()`)
	return err
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
