package security

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

// BCrypt Configuration
// Cost 14 provides military-grade protection against brute-force attacks
// Each increment doubles the computation time
const (
	BCryptMinCost     = 12 // Minimum acceptable cost
	BCryptDefaultCost = 14 // Recommended for financial applications
	BCryptMaxCost     = 16 // Maximum (may cause noticeable delay)
)

var (
	ErrBCryptCostTooLow = errors.New("bcrypt cost must be at least 12")
	ErrEmptyPassword    = errors.New("password cannot be empty")
	ErrPasswordTooLong  = errors.New("password exceeds maximum length of 72 bytes")
)

// HashPasswordBCrypt hashes a password using BCrypt with the specified cost.
// If cost is 0, BCryptDefaultCost (14) is used.
// BCrypt is provided as a fallback option. Argon2id is the preferred algorithm.
func HashPasswordBCrypt(password string, cost int) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}

	// BCrypt has a maximum password length of 72 bytes
	if len(password) > 72 {
		return "", ErrPasswordTooLong
	}

	if cost == 0 {
		cost = BCryptDefaultCost
	}

	if cost < BCryptMinCost {
		return "", ErrBCryptCostTooLow
	}

	if cost > BCryptMaxCost {
		cost = BCryptMaxCost
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

// VerifyPasswordBCrypt verifies a password against a BCrypt hash.
// Returns true if the password matches, false otherwise.
func VerifyPasswordBCrypt(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, ErrEmptyPassword
	}

	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err == bcrypt.ErrMismatchedHashAndPassword {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

// IsBCryptHash checks if a hash string appears to be a BCrypt hash.
// This is useful for auto-detecting hash format during password verification.
func IsBCryptHash(hash string) bool {
	if len(hash) < 4 {
		return false
	}
	// BCrypt hashes start with $2a$, $2b$, or $2y$
	return hash[0] == '$' && hash[1] == '2' && (hash[2] == 'a' || hash[2] == 'b' || hash[2] == 'y') && hash[3] == '$'
}

// IsArgon2Hash checks if a hash string appears to be an Argon2id hash.
// This is useful for auto-detecting hash format during password verification.
func IsArgon2Hash(hash string) bool {
	if len(hash) < 10 {
		return false
	}
	return hash[:9] == "$argon2id"
}

// VerifyPasswordAuto automatically detects the hash format and verifies.
// Supports both Argon2id (preferred) and BCrypt (legacy) hashes.
func VerifyPasswordAuto(password, hash string) (bool, error) {
	if password == "" || hash == "" {
		return false, ErrEmptyPassword
	}

	if IsArgon2Hash(hash) {
		return VerifyPassword(password, hash)
	}

	if IsBCryptHash(hash) {
		return VerifyPasswordBCrypt(password, hash)
	}

	return false, ErrInvalidHash
}
