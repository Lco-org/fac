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
	ErrInvalidKey        = errors.New("encryption key must be 32 bytes for AES-256")
	ErrDecryptionFailed  = errors.New("decryption failed: data may be tampered")
	ErrInvalidCiphertext = errors.New("invalid ciphertext format")
)

type CryptoService struct {
	key []byte
}

func NewCryptoService(key []byte) (*CryptoService, error) {
	if len(key) != 32 {
		return nil, ErrInvalidKey
	}
	keyCopy := make([]byte, 32)
	copy(keyCopy, key)
	return &CryptoService{key: keyCopy}, nil
}

func NewCryptoServiceFromHex(hexKey string) (*CryptoService, error) {
	key, err := decodeHexKey(hexKey)
	if err != nil {
		return nil, err
	}
	return NewCryptoService(key)
}

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

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

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

	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// EncryptString is a convenience method for encrypting string data
func (s *CryptoService) EncryptString(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	return s.Encrypt([]byte(plaintext))
}

// DecryptString is a convenience method for decrypting to string data
func (s *CryptoService) DecryptString(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	plaintext, err := s.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// EncryptSensitiveField encrypts a field only if encryption is enabled (key is set)
// Returns the original value if encryption is not configured
func (s *CryptoService) EncryptSensitiveField(value string) (string, error) {
	if s == nil || len(s.key) == 0 {
		return value, nil
	}
	return s.EncryptString(value)
}

// DecryptSensitiveField decrypts a field only if it appears to be encrypted
// Returns the original value if it doesn't look encrypted
func (s *CryptoService) DecryptSensitiveField(value string) (string, error) {
	if s == nil || len(s.key) == 0 {
		return value, nil
	}
	// Check if the value looks like base64-encoded ciphertext
	if len(value) < 20 {
		return value, nil
	}
	decrypted, err := s.DecryptString(value)
	if err != nil {
		// If decryption fails, assume it's not encrypted
		return value, nil
	}
	return decrypted, nil
}
