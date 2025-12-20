package validation

import (
	"cso-erp/internal/models"
)

// ============================================================================
// AUTH REQUEST VALIDATORS
// ============================================================================

// ValidateLoginRequest validates login request
func ValidateLoginRequest(username, password string) ValidationErrors {
	v := New()

	v.Required("username", username)
	if username != "" {
		v.Email("username", username)
	}

	v.Required("password", password)

	return v.Errors()
}

// Validate2FARequest validates 2FA verification request
func Validate2FARequest(userID, code string) ValidationErrors {
	v := New()

	v.UUIDRequired("user_id", userID)

	v.Required("code", code)
	if code != "" {
		v.LengthBetween("code", code, 6, 6)
		v.AlphaNumeric("code", code)
	}

	return v.Errors()
}

// ValidateChangePasswordRequest validates password change request
func ValidateChangePasswordRequest(oldPassword, newPassword string) ValidationErrors {
	v := New()

	v.Required("old_password", oldPassword)

	v.Required("new_password", newPassword)
	if newPassword != "" {
		v.Password("new_password", newPassword)
	}

	return v.Errors()
}

// ============================================================================
// USER REQUEST VALIDATORS
// ============================================================================

// ValidateCreateUserRequest validates create user request
func ValidateCreateUserRequest(username, email, password, fullName string, role models.UserRole) ValidationErrors {
	v := New()

	v.Required("username", username)
	if username != "" {
		v.MinLength("username", username, 3)
		v.MaxLength("username", username, 50)
		v.AlphaNumericUnderscore("username", username)
	}

	v.Required("email", email)
	if email != "" {
		v.EmailDomain("email", email, []string{"lco.org.ly", "test.lco.org.ly"})
	}

	v.Required("password", password)
	if password != "" {
		v.Password("password", password)
	}

	v.Required("full_name", fullName)
	if fullName != "" {
		v.MaxLength("full_name", fullName, 100)
	}

	validRoles := []string{
		string(models.RoleAdmin),
		string(models.RoleFinance),
		string(models.RoleAuditor),
		string(models.RoleHR),
	}
	v.InEnum("role", string(role), validRoles)

	return v.Errors()
}

// ValidateUpdateUserRequest validates update user request
func ValidateUpdateUserRequest(fullName, fullNameAr string, role models.UserRole) ValidationErrors {
	v := New()

	if fullName != "" {
		v.MaxLength("full_name", fullName, 100)
	}

	if fullNameAr != "" {
		v.MaxLength("full_name_ar", fullNameAr, 100)
	}

	if role != "" {
		validRoles := []string{
			string(models.RoleAdmin),
			string(models.RoleFinance),
			string(models.RoleAuditor),
			string(models.RoleHR),
		}
		v.InEnum("role", string(role), validRoles)
	}

	return v.Errors()
}

// ============================================================================
// TRANSACTION REQUEST VALIDATORS
// ============================================================================

// ValidateCreateTransactionRequest validates create transaction request
func ValidateCreateTransactionRequest(
	txType models.TransactionType,
	amount float64,
	currency models.CurrencyCode,
	exchangeRate float64,
	containerID, projectID, description string,
) ValidationErrors {
	v := New()

	// Validate transaction type
	validTypes := []string{
		string(models.TransactionIncomeGrant),
		string(models.TransactionIncomeDonation),
		string(models.TransactionDisbursementCustody),
		string(models.TransactionDisbursementSalary),
		string(models.TransactionDisbursementOperational),
		string(models.TransactionDisbursementDonation),
		string(models.TransactionDisbursementBonus),
		string(models.TransactionProcurement),
		string(models.TransactionTransfer),
	}
	v.InEnum("type", string(txType), validTypes)

	// Validate amount
	v.PositiveFloat("amount", amount)
	v.FloatRange("amount", amount, 0.01, 10000000000) // Max 10 billion

	// Validate currency
	validCurrencies := []string{
		string(models.CurrencyLYD),
		string(models.CurrencyUSD),
		string(models.CurrencyEUR),
		string(models.CurrencyGBP),
	}
	v.InEnum("currency", string(currency), validCurrencies)

	// Validate exchange rate
	v.PositiveFloat("exchange_rate", exchangeRate)
	v.FloatRange("exchange_rate", exchangeRate, 0.0001, 1000000) // Reasonable range

	// Validate IDs
	v.UUIDRequired("container_id", containerID)
	v.UUIDRequired("project_id", projectID)

	// Validate description
	v.Required("description", description)
	if description != "" {
		v.MaxLength("description", description, 500)
	}

	return v.Errors()
}

// ValidateReverseTransactionRequest validates reverse transaction request
func ValidateReverseTransactionRequest(transactionID, reason string) ValidationErrors {
	v := New()

	v.UUIDRequired("transaction_id", transactionID)

	v.Required("reason", reason)
	if reason != "" {
		v.MinLength("reason", reason, 10)
		v.MaxLength("reason", reason, 500)
	}

	return v.Errors()
}

// ============================================================================
// ORGANIZATION/TENANT REQUEST VALIDATORS
// ============================================================================

// ValidateCreateOrganizationRequest validates create organization request
func ValidateCreateOrganizationRequest(name, domain, contactEmail string) ValidationErrors {
	v := New()

	v.Required("name", name)
	if name != "" {
		v.MinLength("name", name, 2)
		v.MaxLength("name", name, 100)
	}

	v.Required("domain", domain)
	if domain != "" {
		v.Domain("domain", domain)
	}

	v.Required("contact_email", contactEmail)
	if contactEmail != "" {
		v.Email("contact_email", contactEmail)
	}

	return v.Errors()
}

// ValidateVerificationCodeRequest validates verification code request
func ValidateVerificationCodeRequest(orgID, code string) ValidationErrors {
	v := New()

	v.UUIDRequired("organization_id", orgID)

	v.Required("code", code)
	if code != "" {
		v.LengthBetween("code", code, 6, 6)
	}

	return v.Errors()
}

// ============================================================================
// CONTAINER REQUEST VALIDATORS
// ============================================================================

// ValidateCreateContainerRequest validates create container request
func ValidateCreateContainerRequest(name, containerType, currency string) ValidationErrors {
	v := New()

	v.Required("name", name)
	if name != "" {
		v.MinLength("name", name, 2)
		v.MaxLength("name", name, 100)
	}

	v.Required("type", containerType)
	v.InEnum("type", containerType, []string{"cash", "bank"})

	v.Required("currency", currency)
	v.InEnum("currency", currency, []string{"LYD", "USD", "EUR", "GBP"})

	return v.Errors()
}

// ============================================================================
// PROJECT REQUEST VALIDATORS
// ============================================================================

// ValidateCreateProjectRequest validates create project request
func ValidateCreateProjectRequest(code, name, currency string, budget float64) ValidationErrors {
	v := New()

	v.Required("code", code)
	if code != "" {
		v.MinLength("code", code, 2)
		v.MaxLength("code", code, 20)
		v.AlphaNumericUnderscore("code", code)
	}

	v.Required("name", name)
	if name != "" {
		v.MinLength("name", name, 2)
		v.MaxLength("name", name, 200)
	}

	v.Required("currency", currency)
	v.InEnum("currency", currency, []string{"LYD", "USD", "EUR", "GBP"})

	if budget != 0 {
		v.NonNegativeFloat("budget", budget)
	}

	return v.Errors()
}

// ============================================================================
// EMPLOYEE/MEMBER REQUEST VALIDATORS
// ============================================================================

// ValidateCreateEmployeeRequest validates create employee request
func ValidateCreateEmployeeRequest(name, nameAr string, projectID string) ValidationErrors {
	v := New()

	v.Required("name", name)
	if name != "" {
		v.MaxLength("name", name, 100)
	}

	if nameAr != "" {
		v.MaxLength("name_ar", nameAr, 100)
	}

	if projectID != "" {
		v.UUID("project_id", projectID)
	}

	return v.Errors()
}

// ValidateCreateMemberRequest validates create member request
func ValidateCreateMemberRequest(name, memberType string, fee float64) ValidationErrors {
	v := New()

	v.Required("name", name)
	if name != "" {
		v.MaxLength("name", name, 100)
	}

	v.Required("type", memberType)
	v.InEnum("type", memberType, []string{"annual", "lifetime", "honorary"})

	if fee != 0 {
		v.NonNegativeFloat("fee", fee)
	}

	return v.Errors()
}
