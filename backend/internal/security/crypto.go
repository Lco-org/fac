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
