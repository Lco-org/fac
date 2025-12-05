package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

var (
	// ErrInvalidKey indicates the encryption key is not 32 bytes for AES-256
	ErrInvalidKey = errors.New("encryption key must be 32 bytes for AES-256")
	// ErrDecryptionFailed indicates decryption failure, possibly due to tampering
	ErrDecryptionFailed = errors.New("decryption failed: data may be tampered")
	// ErrInvalidCiphertext indicates the ciphertext format is invalid
	ErrInvalidCiphertext = errors.New("invalid ciphertext format")
)

// CryptoService handles AES-256-GCM encryption/decryption
// AES-256-GCM provides authenticated encryption with associated data (AEAD)
// This ensures both confidentiality and integrity of encrypted data
type CryptoService struct {
	key []byte
}

// NewCryptoService creates a new encryption service with the given 32-byte key
// The key should be cryptographically random and securely stored
func NewCryptoService(key []byte) (*CryptoService, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	// Create a copy of the key to prevent external modifications
	keyCopy := make([]byte, 32)
	copy(keyCopy, key)
	return &CryptoService{key: keyCopy}, nil
}

// NewCryptoServiceFromHex creates a CryptoService from a hex-encoded key string
func NewCryptoServiceFromHex(hexKey string) (*CryptoService, error) {
	key, err := decodeHexKey(hexKey)
	if err != nil {
		return nil, err
	}
	return NewCryptoService(key)
}

// decodeHexKey decodes a hex string to bytes
func decodeHexKey(hexKey string) ([]byte, error) {
	if len(hexKey) != 64 {
		return nil, ErrInvalidKey
	}
	key := make([]byte, 32)
	for i := 0; i < 32; i++ {
		var b byte
		for j := 0; j < 2; j++ {
			c := hexKey[i*2+j]
			switch {
			case c >= '0' && c <= '9':
				b = b*16 + (c - '0')
			case c >= 'a' && c <= 'f':
				b = b*16 + (c - 'a' + 10)
			case c >= 'A' && c <= 'F':
				b = b*16 + (c - 'A' + 10)
			default:
				return nil, ErrInvalidKey
			}
		}
		key[i] = b
	}
	return key, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
// Returns base64-encoded ciphertext with prepended nonce
// Each encryption uses a unique cryptographically random nonce
func (s *CryptoService) Encrypt(plaintext []byte) (string, error) {
	if len(plaintext) == 0 {
		return "", nil
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create unique nonce for each encryption (12 bytes for GCM)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and prepend nonce to ciphertext
	// GCM mode provides authentication tag automatically
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
// Returns error if data has been tampered with
func (s *CryptoService) Decrypt(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, ErrInvalidCiphertext
	}

	// Extract nonce and decrypt
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	// GCM Open will verify the authentication tag
	// If tampering is detected, it returns an error
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns base64 encoded result
func (s *CryptoService) EncryptString(plaintext string) (string, error) {
	return s.Encrypt([]byte(plaintext))
}

// DecryptString decrypts and returns the original string
func (s *CryptoService) DecryptString(encoded string) (string, error) {
	plaintext, err := s.Decrypt(encoded)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptFloat64 encrypts a float64 value for secure storage
func (s *CryptoService) EncryptFloat64(value float64) (string, error) {
	// Convert float64 to string with full precision
	str := formatFloat64(value)
	return s.EncryptString(str)
}

// DecryptFloat64 decrypts an encrypted float64 value
func (s *CryptoService) DecryptFloat64(encoded string) (float64, error) {
	str, err := s.DecryptString(encoded)
	if err != nil {
		return 0, err
	}
	return parseFloat64(str)
}

// formatFloat64 converts float64 to string with full precision
func formatFloat64(f float64) string {
	// Use enough precision to preserve the value
	return string(appendFloat64(nil, f))
}

// parseFloat64 parses a string to float64
func parseFloat64(s string) (float64, error) {
	var result float64
	var negative bool
	var i int

	if len(s) == 0 {
		return 0, errors.New("empty string")
	}

	if s[0] == '-' {
		negative = true
		i++
	}

	// Parse integer part
	for ; i < len(s) && s[i] != '.'; i++ {
		if s[i] < '0' || s[i] > '9' {
			return 0, errors.New("invalid number")
		}
		result = result*10 + float64(s[i]-'0')
	}

	// Parse decimal part
	if i < len(s) && s[i] == '.' {
		i++
		var decimal float64
		var divisor float64 = 10
		for ; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				return 0, errors.New("invalid number")
			}
			decimal += float64(s[i]-'0') / divisor
			divisor *= 10
		}
		result += decimal
	}

	if negative {
		result = -result
	}

	return result, nil
}

// appendFloat64 appends float64 to byte slice
func appendFloat64(dst []byte, f float64) []byte {
	if f < 0 {
		dst = append(dst, '-')
		f = -f
	}

	// Handle integer part
	intPart := int64(f)
	fracPart := f - float64(intPart)

	// Convert integer part
	if intPart == 0 {
		dst = append(dst, '0')
	} else {
		var digits [20]byte
		n := 0
		for intPart > 0 {
			digits[n] = byte(intPart%10) + '0'
			intPart /= 10
			n++
		}
		for i := n - 1; i >= 0; i-- {
			dst = append(dst, digits[i])
		}
	}

	// Handle fractional part with up to 10 decimal places
	if fracPart > 0 {
		dst = append(dst, '.')
		for i := 0; i < 10 && fracPart > 1e-10; i++ {
			fracPart *= 10
			digit := int(fracPart)
			dst = append(dst, byte(digit)+'0')
			fracPart -= float64(digit)
		}
	}

	return dst
}
