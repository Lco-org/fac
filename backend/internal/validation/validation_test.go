package validation

import (
	"cso-erp/internal/models"
	"testing"
)

// ============================================================================
// VALIDATOR TESTS
// ============================================================================

func TestValidator_Required(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"Non-empty", "hello", false},
		{"Empty", "", true},
		{"Whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Required("field", tt.value)
			if tt.wantErr && !v.HasErrors() {
				t.Error("Expected error, got none")
			}
			if !tt.wantErr && v.HasErrors() {
				t.Errorf("Expected no error, got: %v", v.Errors())
			}
		})
	}
}

func TestValidator_MinLength(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		min     int
		wantErr bool
	}{
		{"Exact length", "abc", 3, false},
		{"Above min", "abcd", 3, false},
		{"Below min", "ab", 3, true},
		{"Unicode", "مرحبا", 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.MinLength("field", tt.value, tt.min)
			if tt.wantErr && !v.HasErrors() {
				t.Error("Expected error, got none")
			}
			if !tt.wantErr && v.HasErrors() {
				t.Errorf("Expected no error, got: %v", v.Errors())
			}
		})
	}
}

func TestValidator_Email(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"user@example.com", false},
		{"test@lco.org.ly", false},
		{"invalid", true},
		{"@example.com", true},
		{"user@", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			v := New()
			v.Email("email", tt.email)
			if tt.wantErr && !v.HasErrors() {
				t.Error("Expected error, got none")
			}
			if !tt.wantErr && v.HasErrors() {
				t.Errorf("Expected no error, got: %v", v.Errors())
			}
		})
	}
}

func TestValidator_Password(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"Valid", "Password123!", false},
		{"Too short", "Pass1!", true},
		{"Exactly 8", "Pass123!", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Password("password", tt.password)
			if tt.wantErr && !v.HasErrors() {
				t.Error("Expected error, got none")
			}
			if !tt.wantErr && v.HasErrors() {
				t.Errorf("Expected no error, got: %v", v.Errors())
			}
		})
	}
}

func TestValidator_UUID(t *testing.T) {
	tests := []struct {
		uuid    string
		wantErr bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid-uuid", true},
		{"", true},
		{"550e8400-e29b-41d4-a716", true},
	}

	for _, tt := range tests {
		t.Run(tt.uuid, func(t *testing.T) {
			v := New()
			v.UUID("id", tt.uuid)
			if tt.wantErr != v.HasErrors() {
				t.Errorf("UUID(%q): wantErr=%v, hasErrors=%v", tt.uuid, tt.wantErr, v.HasErrors())
			}
		})
	}
}

func TestValidator_PositiveFloat(t *testing.T) {
	tests := []struct {
		value   float64
		wantErr bool
	}{
		{100.00, false},
		{0.01, false},
		{0, true},
		{-1, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			v := New()
			v.PositiveFloat("amount", tt.value)
			if tt.wantErr != v.HasErrors() {
				t.Errorf("PositiveFloat(%v): wantErr=%v, hasErrors=%v", tt.value, tt.wantErr, v.HasErrors())
			}
		})
	}
}

func TestValidator_InEnum(t *testing.T) {
	allowed := []string{"admin", "finance", "auditor", "hr"}

	tests := []struct {
		value   string
		wantErr bool
	}{
		{"admin", false},
		{"finance", false},
		{"unknown", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			v := New()
			v.InEnum("role", tt.value, allowed)
			if tt.wantErr != v.HasErrors() {
				t.Errorf("InEnum(%q): wantErr=%v, hasErrors=%v", tt.value, tt.wantErr, v.HasErrors())
			}
		})
	}
}

func TestValidator_Domain(t *testing.T) {
	tests := []struct {
		domain  string
		wantErr bool
	}{
		{"example.com", false},
		{"sub.example.com", false},
		{"lco.org.ly", false},
		{"invalid", true},
		{".com", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.domain, func(t *testing.T) {
			v := New()
			v.Domain("domain", tt.domain)
			if tt.wantErr != v.HasErrors() {
				t.Errorf("Domain(%q): wantErr=%v, hasErrors=%v", tt.domain, tt.wantErr, v.HasErrors())
			}
		})
	}
}

// ============================================================================
// REQUEST VALIDATOR TESTS
// ============================================================================

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		wantErr  bool
	}{
		{"Valid", "user@lco.org.ly", "Password123!", false},
		{"Missing username", "", "Password123!", true},
		{"Missing password", "user@lco.org.ly", "", true},
		{"Invalid email", "invalid", "Password123!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateLoginRequest(tt.username, tt.password)
			if tt.wantErr && !errs.HasErrors() {
				t.Error("Expected errors, got none")
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("Expected no errors, got: %v", errs)
			}
		})
	}
}

func TestValidateCreateUserRequest(t *testing.T) {
	tests := []struct {
		name     string
		username string
		email    string
		password string
		fullName string
		role     models.UserRole
		wantErr  bool
	}{
		{"Valid", "testuser", "user@lco.org.ly", "Password123!", "Test User", models.RoleAdmin, false},
		{"Short username", "ab", "user@lco.org.ly", "Password123!", "Test User", models.RoleAdmin, true},
		{"Invalid email domain", "testuser", "user@gmail.com", "Password123!", "Test User", models.RoleAdmin, true},
		{"Weak password", "testuser", "user@lco.org.ly", "short", "Test User", models.RoleAdmin, true},
		{"Invalid role", "testuser", "user@lco.org.ly", "Password123!", "Test User", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateUserRequest(tt.username, tt.email, tt.password, tt.fullName, tt.role)
			if tt.wantErr && !errs.HasErrors() {
				t.Error("Expected errors, got none")
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("Expected no errors, got: %v", errs)
			}
		})
	}
}

func TestValidateCreateTransactionRequest(t *testing.T) {
	validUUID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name         string
		txType       models.TransactionType
		amount       float64
		currency     models.CurrencyCode
		exchangeRate float64
		containerID  string
		projectID    string
		description  string
		wantErr      bool
	}{
		{"Valid", models.TransactionIncomeGrant, 1000.00, models.CurrencyLYD, 1.0, validUUID, validUUID, "Test transaction", false},
		{"Zero amount", models.TransactionIncomeGrant, 0, models.CurrencyLYD, 1.0, validUUID, validUUID, "Test", true},
		{"Negative amount", models.TransactionIncomeGrant, -100, models.CurrencyLYD, 1.0, validUUID, validUUID, "Test", true},
		{"Invalid type", "invalid", 1000.00, models.CurrencyLYD, 1.0, validUUID, validUUID, "Test", true},
		{"Invalid UUID", models.TransactionIncomeGrant, 1000.00, models.CurrencyLYD, 1.0, "invalid", validUUID, "Test", true},
		{"Empty description", models.TransactionIncomeGrant, 1000.00, models.CurrencyLYD, 1.0, validUUID, validUUID, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateTransactionRequest(tt.txType, tt.amount, tt.currency, tt.exchangeRate, tt.containerID, tt.projectID, tt.description)
			if tt.wantErr && !errs.HasErrors() {
				t.Error("Expected errors, got none")
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("Expected no errors, got: %v", errs)
			}
		})
	}
}

func TestValidateCreateOrganizationRequest(t *testing.T) {
	tests := []struct {
		name         string
		orgName      string
		domain       string
		contactEmail string
		wantErr      bool
	}{
		{"Valid", "Acme Corp", "acme.com", "admin@acme.com", false},
		{"Missing name", "", "acme.com", "admin@acme.com", true},
		{"Invalid domain", "Acme Corp", "invalid", "admin@acme.com", true},
		{"Invalid email", "Acme Corp", "acme.com", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ValidateCreateOrganizationRequest(tt.orgName, tt.domain, tt.contactEmail)
			if tt.wantErr && !errs.HasErrors() {
				t.Error("Expected errors, got none")
			}
			if !tt.wantErr && errs.HasErrors() {
				t.Errorf("Expected no errors, got: %v", errs)
			}
		})
	}
}
