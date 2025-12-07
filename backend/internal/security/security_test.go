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

// =====================================================
// BCrypt Tests
// =====================================================

func TestBCrypt_HashAndVerify(t *testing.T) {
	password := "SecurePassword123!"

	hash, err := HashPasswordBCrypt(password, BCryptDefaultCost)
	if err != nil {
		t.Fatalf("HashPasswordBCrypt failed: %v", err)
	}

	if !IsBCryptHash(hash) {
		t.Error("Hash should be identified as BCrypt")
	}

	valid, err := VerifyPasswordBCrypt(password, hash)
	if err != nil {
		t.Fatalf("VerifyPasswordBCrypt failed: %v", err)
	}
	if !valid {
		t.Error("Correct password should verify successfully")
	}

	valid, err = VerifyPasswordBCrypt("wrongpassword", hash)
	if err != nil {
		t.Fatalf("VerifyPasswordBCrypt failed: %v", err)
	}
	if valid {
		t.Error("Wrong password should not verify")
	}
}

func TestBCrypt_Cost(t *testing.T) {
	password := "TestPassword123"

	// Test minimum cost
	_, err := HashPasswordBCrypt(password, 10)
	if err != ErrBCryptCostTooLow {
		t.Error("Cost below 12 should return ErrBCryptCostTooLow")
	}

	// Test valid costs
	for _, cost := range []int{12, 13, 14} {
		hash, err := HashPasswordBCrypt(password, cost)
		if err != nil {
			t.Errorf("Cost %d should work: %v", cost, err)
		}
		valid, _ := VerifyPasswordBCrypt(password, hash)
		if !valid {
			t.Errorf("Password should verify with cost %d", cost)
		}
	}
}

func TestBCrypt_EmptyPassword(t *testing.T) {
	_, err := HashPasswordBCrypt("", 14)
	if err != ErrEmptyPassword {
		t.Error("Empty password should return ErrEmptyPassword")
	}
}

func TestBCrypt_LongPassword(t *testing.T) {
	longPassword := string(make([]byte, 73)) // 73 bytes, exceeds 72 limit
	_, err := HashPasswordBCrypt(longPassword, 14)
	if err != ErrPasswordTooLong {
		t.Error("Password over 72 bytes should return ErrPasswordTooLong")
	}
}

// =====================================================
// Hash Auto-Detection Tests
// =====================================================

func TestHashAutoDetection(t *testing.T) {
	password := "TestPassword123!"

	// Test Argon2id detection
	argon2Hash, _ := HashPassword(password, nil)
	if !IsArgon2Hash(argon2Hash) {
		t.Error("Argon2id hash should be detected")
	}
	if IsBCryptHash(argon2Hash) {
		t.Error("Argon2id hash should not be detected as BCrypt")
	}

	// Test BCrypt detection
	bcryptHash, _ := HashPasswordBCrypt(password, 14)
	if !IsBCryptHash(bcryptHash) {
		t.Error("BCrypt hash should be detected")
	}
	if IsArgon2Hash(bcryptHash) {
		t.Error("BCrypt hash should not be detected as Argon2id")
	}

	// Test auto-verification
	validArgon2, err := VerifyPasswordAuto(password, argon2Hash)
	if err != nil || !validArgon2 {
		t.Error("Auto-verify should work with Argon2id hash")
	}

	validBCrypt, err := VerifyPasswordAuto(password, bcryptHash)
	if err != nil || !validBCrypt {
		t.Error("Auto-verify should work with BCrypt hash")
	}

	// Test wrong password
	invalidArgon2, _ := VerifyPasswordAuto("wrong", argon2Hash)
	if invalidArgon2 {
		t.Error("Wrong password should not verify")
	}
}

// =====================================================
// Input Sanitizer Tests
// =====================================================

func TestInputSanitizer_SQLInjection(t *testing.T) {
	sanitizer := NewInputSanitizer()

	injectionPayloads := []string{
		"' OR '1'='1",
		"1; DROP TABLE users--",
		"UNION SELECT * FROM users",
		"1' AND SLEEP(5)--",
		"admin'--",
	}

	for _, payload := range injectionPayloads {
		if !sanitizer.ContainsSQLInjection(payload) {
			t.Errorf("Should detect SQL injection: %q", payload)
		}
	}

	safeInputs := []string{
		"John Doe",
		"user@example.com",
		"Hello World 123",
		"مرحبا بالعالم",
	}

	for _, input := range safeInputs {
		if sanitizer.ContainsSQLInjection(input) {
			t.Errorf("Should not flag safe input: %q", input)
		}
	}
}

func TestInputSanitizer_XSS(t *testing.T) {
	sanitizer := NewInputSanitizer()

	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
	}

	for _, payload := range xssPayloads {
		if !sanitizer.ContainsXSS(payload) {
			t.Errorf("Should detect XSS: %q", payload)
		}
	}
}

func TestInputSanitizer_UUID(t *testing.T) {
	sanitizer := NewInputSanitizer()

	validUUIDs := []string{
		"550e8400-e29b-41d4-a716-446655440000",
		"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
	}

	for _, uuid := range validUUIDs {
		result := sanitizer.SanitizeUUID(uuid)
		if result == "" {
			t.Errorf("Valid UUID should pass: %q", uuid)
		}
	}

	invalidUUIDs := []string{
		"not-a-uuid",
		"550e8400-e29b-41d4-a716",
		"'; DROP TABLE users--",
	}

	for _, uuid := range invalidUUIDs {
		result := sanitizer.SanitizeUUID(uuid)
		if result != "" {
			t.Errorf("Invalid UUID should fail: %q", uuid)
		}
	}
}

func TestInputSanitizer_FinancialAmount(t *testing.T) {
	sanitizer := NewInputSanitizer()

	validAmounts := []float64{0.01, 100.50, 999999.99}
	for _, amount := range validAmounts {
		if !sanitizer.ValidateFinancialAmount(amount) {
			t.Errorf("Valid amount should pass: %f", amount)
		}
	}

	invalidAmounts := []float64{0, -100, 10000000001}
	for _, amount := range invalidAmounts {
		if sanitizer.ValidateFinancialAmount(amount) {
			t.Errorf("Invalid amount should fail: %f", amount)
		}
	}
}

// =====================================================
// BCrypt Benchmark
// =====================================================

func BenchmarkBCryptHash(b *testing.B) {
	password := "BenchmarkPassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		HashPasswordBCrypt(password, BCryptDefaultCost)
	}
}
