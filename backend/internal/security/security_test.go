package security

import (
	"bytes"
	"testing"
)

func TestCryptoService_EncryptDecrypt(t *testing.T) {
	// Create a valid 32-byte key
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
		plaintext string
	}{
		{"empty string", ""},
		{"simple text", "Hello, World!"},
		{"arabic text", "مرحبا بالعالم"},
		{"financial data", "Balance: 1234567.89 LYD"},
		{"json data", `{"amount": 100.50, "currency": "LYD"}`},
		{"long text", "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Encrypt
			encrypted, err := crypto.EncryptString(tt.plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Empty string returns empty
			if tt.plaintext == "" && encrypted == "" {
				return
			}

			// Encrypted should be different from plaintext
			if encrypted == tt.plaintext {
				t.Error("Encrypted text should be different from plaintext")
			}

			// Decrypt
			decrypted, err := crypto.DecryptString(encrypted)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Decrypted should match original
			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text mismatch: got %q, want %q", decrypted, tt.plaintext)
			}
		})
	}
}

func TestCryptoService_UniqueNonce(t *testing.T) {
	key := make([]byte, 32)
	crypto, _ := NewCryptoService(key)

	plaintext := "test data"

	// Encrypt same data multiple times
	encrypted1, _ := crypto.EncryptString(plaintext)
	encrypted2, _ := crypto.EncryptString(plaintext)
	encrypted3, _ := crypto.EncryptString(plaintext)

	// Each encryption should produce different ciphertext (due to unique nonce)
	if encrypted1 == encrypted2 || encrypted2 == encrypted3 || encrypted1 == encrypted3 {
		t.Error("Each encryption should produce unique ciphertext")
	}

	// But all should decrypt to same plaintext
	d1, _ := crypto.DecryptString(encrypted1)
	d2, _ := crypto.DecryptString(encrypted2)
	d3, _ := crypto.DecryptString(encrypted3)

	if d1 != plaintext || d2 != plaintext || d3 != plaintext {
		t.Error("All decryptions should produce same plaintext")
	}
}

func TestCryptoService_TamperDetection(t *testing.T) {
	key := make([]byte, 32)
	crypto, _ := NewCryptoService(key)

	plaintext := "sensitive financial data"
	encrypted, _ := crypto.EncryptString(plaintext)

	// Tamper with the encrypted data
	tampered := encrypted[:len(encrypted)-2] + "XX"

	_, err := crypto.DecryptString(tampered)
	if err != ErrDecryptionFailed && err != ErrInvalidCiphertext {
		t.Error("Tampered data should fail decryption")
	}
}

func TestCryptoService_InvalidKey(t *testing.T) {
	// Test with wrong key sizes
	invalidSizes := []int{0, 16, 24, 31, 33, 64}

	for _, size := range invalidSizes {
		key := make([]byte, size)
		_, err := NewCryptoService(key)
		if err != ErrInvalidKey {
			t.Errorf("Key size %d should return ErrInvalidKey", size)
		}
	}
}

func TestCryptoService_Float64(t *testing.T) {
	key := make([]byte, 32)
	crypto, _ := NewCryptoService(key)

	tests := []float64{
		0,
		1.5,
		-100.25,
		1234567.89,
		0.001,
		99999999.99999999,
	}

	for _, val := range tests {
		encrypted, err := crypto.EncryptFloat64(val)
		if err != nil {
			t.Fatalf("Failed to encrypt float64 %f: %v", val, err)
		}

		decrypted, err := crypto.DecryptFloat64(encrypted)
		if err != nil {
			t.Fatalf("Failed to decrypt float64: %v", err)
		}

		// Allow small precision difference
		diff := val - decrypted
		if diff < 0 {
			diff = -diff
		}
		if diff > 0.0000001 {
			t.Errorf("Float64 mismatch: got %f, want %f", decrypted, val)
		}
	}
}

func TestArgon2_HashAndVerify(t *testing.T) {
	password := "SecurePassword123!"

	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Hash should have correct format
	if len(hash) < 50 {
		t.Error("Hash seems too short")
	}

	// Verify correct password
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !valid {
		t.Error("Correct password should verify successfully")
	}

	// Verify wrong password
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

	// Each hash should be unique (different salt)
	if hash1 == hash2 {
		t.Error("Same password should produce different hashes (different salts)")
	}

	// Both should verify correctly
	valid1, _ := VerifyPassword(password, hash1)
	valid2, _ := VerifyPassword(password, hash2)

	if !valid1 || !valid2 {
		t.Error("Both hashes should verify with correct password")
	}
}

func TestArgon2_CustomParams(t *testing.T) {
	password := "TestPassword"

	params := &Argon2Params{
		Memory:      32 * 1024, // 32MB
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
		// Valid emails
		{"user@lco.org.ly", nil},
		{"admin@lco.org.ly", nil},
		{"finance.dept@lco.org.ly", nil},
		{"user123@lco.org.ly", nil},

		// Invalid domain
		{"user@gmail.com", ErrUnauthorizedDomain},
		{"user@company.com", ErrUnauthorizedDomain},
		{"user@lco.org", ErrUnauthorizedDomain},

		// Invalid format
		{"", ErrInvalidEmailFormat},
		{"notanemail", ErrInvalidEmailFormat},
		{"@lco.org.ly", ErrInvalidEmailFormat},
		{"user@", ErrInvalidEmailFormat},
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
	// Test that the hash calculation is deterministic
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

	// Modify entry and verify hash changes
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

	// Test nil map
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
