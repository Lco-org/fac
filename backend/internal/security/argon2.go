package security

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	// ErrInvalidHash indicates the password hash format is invalid
	ErrInvalidHash = errors.New("invalid password hash format")
	// ErrIncompatibleVersion indicates the argon2 version is incompatible
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// Argon2Params defines the parameters for Argon2id hashing
// These parameters control the computational cost of hashing
type Argon2Params struct {
	Memory      uint32 // Memory in KB (default: 64*1024 = 64MB)
	Iterations  uint32 // Time cost (default: 3)
	Parallelism uint8  // CPU parallelism (default: 4)
	SaltLength  uint32 // Salt length in bytes (default: 16)
	KeyLength   uint32 // Hash output length (default: 32)
}

// DefaultArgon2Params returns secure default parameters
// These values follow OWASP recommendations for sensitive systems
// Memory-hardness prevents GPU/ASIC-based attacks
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      64 * 1024, // 64MB - high memory to resist GPU attacks
		Iterations:  3,         // 3 iterations - balanced security/performance
		Parallelism: 4,         // 4 threads
		SaltLength:  16,        // 128-bit salt
		KeyLength:   32,        // 256-bit hash
	}
}

// HashPassword creates an Argon2id hash of the password
// Argon2id is the recommended variant, combining resistance to both
// side-channel attacks (Argon2i) and GPU cracking (Argon2d)
func HashPassword(password string, params *Argon2Params) (string, error) {
	if params == nil {
		params = DefaultArgon2Params()
	}

	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Generate cryptographic random salt
	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Generate hash using Argon2id
	// Argon2id provides optimal protection against both timing attacks
	// and memory tradeoff attacks
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Encode as: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	// This format is compatible with most password verification libraries
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		params.Memory,
		params.Iterations,
		params.Parallelism,
		b64Salt,
		b64Hash,
	), nil
}

// VerifyPassword checks if password matches the hash
// Uses constant-time comparison to prevent timing attacks
func VerifyPassword(password, encodedHash string) (bool, error) {
	if password == "" || encodedHash == "" {
		return false, ErrInvalidHash
	}

	params, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return false, err
	}

	// Generate hash with same parameters
	otherHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	// Constant-time comparison to prevent timing attacks
	// This ensures the comparison takes the same time regardless
	// of how many bytes match
	return subtle.ConstantTimeCompare(hash, otherHash) == 1, nil
}

// decodeHash parses the encoded hash string
func decodeHash(encodedHash string) (*Argon2Params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	if parts[1] != "argon2id" {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	params := &Argon2Params{}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d",
		&params.Memory, &params.Iterations, &params.Parallelism); err != nil {
		return nil, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	params.SaltLength = uint32(len(salt))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, ErrInvalidHash
	}
	params.KeyLength = uint32(len(hash))

	return params, salt, hash, nil
}

// NeedsRehash checks if the hash needs to be updated with new parameters
// This is useful when upgrading security parameters over time
func NeedsRehash(encodedHash string, targetParams *Argon2Params) bool {
	if targetParams == nil {
		targetParams = DefaultArgon2Params()
	}

	params, _, _, err := decodeHash(encodedHash)
	if err != nil {
		return true // Invalid hash should be rehashed
	}

	return params.Memory < targetParams.Memory ||
		params.Iterations < targetParams.Iterations ||
		params.Parallelism < targetParams.Parallelism
}
