package security

import (
	"bytes"
	"testing"
)

func TestCryptoService_EncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	crypto, err := NewCryptoService(key)
	if err != nil {
		t.Fatalf("Failed to create crypto service: %v", err)
	}

	tests := []struct {
		name      string
		plaintext []byte
	}{
		{"empty", []byte{}},
		{"simple text", []byte("Hello, World!")},
		{"arabic text", []byte("مرحبا بالعالم")},
		{"binary data", []byte{0x00, 0x01, 0x02, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := crypto.Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			if len(tt.plaintext) == 0 && encrypted == "" {
				return
			}

			decrypted, err := crypto.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if !bytes.Equal(decrypted, tt.plaintext) {
				t.Errorf("Decrypted data mismatch")
			}
		})
	}
}

func TestCryptoService_UniqueNonce(t *testing.T) {
	key := make([]byte, 32)
	crypto, _ := NewCryptoService(key)

	plaintext := []byte("test data")

	encrypted1, _ := crypto.Encrypt(plaintext)
	encrypted2, _ := crypto.Encrypt(plaintext)

	if encrypted1 == encrypted2 {
		t.Error("Each encryption should produce unique ciphertext")
	}

	d1, _ := crypto.Decrypt(encrypted1)
	d2, _ := crypto.Decrypt(encrypted2)

	if !bytes.Equal(d1, plaintext) || !bytes.Equal(d2, plaintext) {
		t.Error("All decryptions should produce same plaintext")
	}
}

func TestCryptoService_TamperDetection(t *testing.T) {
	key := make([]byte, 32)
	crypto, _ := NewCryptoService(key)

	plaintext := []byte("sensitive data")
	encrypted, _ := crypto.Encrypt(plaintext)

	tampered := encrypted[:len(encrypted)-2] + "XX"

	_, err := crypto.Decrypt(tampered)
	if err != ErrDecryptionFailed && err != ErrInvalidCiphertext {
		t.Error("Tampered data should fail decryption")
	}
}

func TestCryptoService_InvalidKey(t *testing.T) {
	invalidSizes := []int{0, 16, 24, 31, 33, 64}

	for _, size := range invalidSizes {
		key := make([]byte, size)
		_, err := NewCryptoService(key)
		if err != ErrInvalidKey {
			t.Errorf("Key size %d should return ErrInvalidKey", size)
		}
	}
}

func TestArgon2_HashAndVerify(t *testing.T) {
	password := "SecurePassword123!"

	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if len(hash) < 50 {
		t.Error("Hash seems too short")
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("Correct password should verify successfully")
	}

	valid, err = VerifyPassword("wrongpassword", hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if valid {
		t.Error("Wrong password should not verify")
	}
}

func TestArgon2_UniqueHash(t *testing.T) {
	password := "SamePassword123"

	hash1, _ := HashPassword(password, nil)
	hash2, _ := HashPassword(password, nil)

	if hash1 == hash2 {
		t.Error("Same password should produce different hashes (different salts)")
	}

	valid1, _ := VerifyPassword(password, hash1)
	valid2, _ := VerifyPassword(password, hash2)

	if !valid1 || !valid2 {
		t.Error("Both hashes should verify with correct password")
	}
}

func TestArgon2_CustomParams(t *testing.T) {
	password := "TestPassword"

	params := &Argon2Params{
		Memory:      32 * 1024,
		Iterations:  2,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("HashPassword with custom params failed: %v", err)
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("Password should verify with custom params")
	}
}

func TestArgon2_EmptyPassword(t *testing.T) {
	_, err := HashPassword("", nil)
	if err == nil {
		t.Error("Empty password should return error")
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr error
	}{
		{"user@lco.org.ly", nil},
		{"admin@lco.org.ly", nil},
		{"finance.dept@lco.org.ly", nil},
		{"user123@lco.org.ly", nil},
		{"user@gmail.com", ErrUnauthorizedDomain},
		{"user@company.com", ErrUnauthorizedDomain},
		{"", ErrInvalidEmailFormat},
		{"notanemail", ErrInvalidEmailFormat},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err != tt.wantErr {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"USER@LCO.ORG.LY", "user@lco.org.ly"},
		{"  user@lco.org.ly  ", "user@lco.org.ly"},
		{"Admin@LCO.org.ly", "admin@lco.org.ly"},
	}

	for _, tt := range tests {
		result := NormalizeEmail(tt.input)
		if result != tt.expected {
			t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestAuditChainedHashing(t *testing.T) {
	entry := &AuditLog{
		Sequence:     1,
		Action:       ActionLogin,
		ResourceType: "auth",
		Details:      `{"test": "data"}`,
		PreviousHash: "GENESIS",
	}
	entry.CreatedAt = entry.CreatedAt.UTC()

	service := &AuditService{}

	hash1 := service.calculateHash(entry)
	hash2 := service.calculateHash(entry)

	if hash1 != hash2 {
		t.Error("Same entry should produce same hash")
	}

	entry.Details = `{"test": "modified"}`
	hash3 := service.calculateHash(entry)

	if hash1 == hash3 {
		t.Error("Modified entry should produce different hash")
	}
}

func TestAuditDetails(t *testing.T) {
	details := AuditDetails(map[string]interface{}{
		"action": "test",
		"amount": 100.5,
	})

	if details == "" || details == "{}" {
		t.Error("AuditDetails should produce non-empty JSON")
	}

	nilDetails := AuditDetails(nil)
	if nilDetails != "{}" {
		t.Error("Nil map should produce empty JSON object")
	}
}

func BenchmarkArgon2Hash(b *testing.B) {
	password := "BenchmarkPassword123"
	params := DefaultArgon2Params()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashPassword(password, params)
	}
}

func BenchmarkAESEncrypt(b *testing.B) {
	key := make([]byte, 32)
	crypto, _ := NewCryptoService(key)
	data := bytes.Repeat([]byte("benchmark data "), 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		crypto.Encrypt(data)
	}
}
