// Package validation provides request validation utilities for the CSO-ERP API
package validation

import (
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// HasErrors checks if there are any validation errors
func (e ValidationErrors) HasErrors() bool {
	return len(e) > 0
}

// ToMap converts errors to a map for JSON responses
func (e ValidationErrors) ToMap() map[string]interface{} {
	result := map[string]interface{}{
		"error":  "Validation failed",
		"errors": e,
	}
	return result
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new Validator instance
func New() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// AddError adds a validation error
func (v *Validator) AddError(field, message, code string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
		Code:    code,
	})
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// HasErrors checks if there are any validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Reset clears all validation errors
func (v *Validator) Reset() {
	v.errors = make(ValidationErrors, 0)
}

// ============================================================================
// STRING VALIDATORS
// ============================================================================

// Required checks if a string is not empty
func (v *Validator) Required(field, value string) bool {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "This field is required", "required")
		return false
	}
	return true
}

// MinLength checks if a string has minimum length
func (v *Validator) MinLength(field, value string, min int) bool {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, fmt.Sprintf("Must be at least %d characters", min), "min_length")
		return false
	}
	return true
}

// MaxLength checks if a string does not exceed maximum length
func (v *Validator) MaxLength(field, value string, max int) bool {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, fmt.Sprintf("Must not exceed %d characters", max), "max_length")
		return false
	}
	return true
}

// LengthBetween checks if string length is within range
func (v *Validator) LengthBetween(field, value string, min, max int) bool {
	length := utf8.RuneCountInString(value)
	if length < min || length > max {
		v.AddError(field, fmt.Sprintf("Must be between %d and %d characters", min, max), "length_between")
		return false
	}
	return true
}

// ============================================================================
// EMAIL VALIDATORS
// ============================================================================

// Email validates email format
func (v *Validator) Email(field, value string) bool {
	_, err := mail.ParseAddress(value)
	if err != nil {
		v.AddError(field, "Invalid email format", "invalid_email")
		return false
	}
	return true
}

// EmailDomain validates email is from allowed domain
func (v *Validator) EmailDomain(field, value string, allowedDomains []string) bool {
	if !v.Email(field, value) {
		return false
	}

	parts := strings.Split(value, "@")
	if len(parts) != 2 {
		return false
	}

	domain := strings.ToLower(parts[1])
	for _, allowed := range allowedDomains {
		if domain == strings.ToLower(allowed) {
			return true
		}
	}

	v.AddError(field, fmt.Sprintf("Email must be from allowed domains: %v", allowedDomains), "invalid_domain")
	return false
}

// ============================================================================
// PASSWORD VALIDATORS
// ============================================================================

// Password validates password strength
func (v *Validator) Password(field, value string) bool {
	valid := true

	if len(value) < 8 {
		v.AddError(field, "Password must be at least 8 characters", "password_too_short")
		valid = false
	}

	if len(value) > 128 {
		v.AddError(field, "Password must not exceed 128 characters", "password_too_long")
		valid = false
	}

	return valid
}

// StrongPassword validates password with complexity requirements
func (v *Validator) StrongPassword(field, value string) bool {
	if !v.Password(field, value) {
		return false
	}

	var (
		hasUpper   = regexp.MustCompile(`[A-Z]`).MatchString(value)
		hasLower   = regexp.MustCompile(`[a-z]`).MatchString(value)
		hasNumber  = regexp.MustCompile(`[0-9]`).MatchString(value)
		hasSpecial = regexp.MustCompile(`[!@#$%^&*(),.?":{}|<>]`).MatchString(value)
	)

	if !hasUpper {
		v.AddError(field, "Password must contain at least one uppercase letter", "password_no_upper")
		return false
	}
	if !hasLower {
		v.AddError(field, "Password must contain at least one lowercase letter", "password_no_lower")
		return false
	}
	if !hasNumber {
		v.AddError(field, "Password must contain at least one number", "password_no_number")
		return false
	}
	if !hasSpecial {
		v.AddError(field, "Password must contain at least one special character", "password_no_special")
		return false
	}

	return true
}

// ============================================================================
// UUID VALIDATORS
// ============================================================================

// UUID validates UUID format
func (v *Validator) UUID(field, value string) bool {
	if _, err := uuid.Parse(value); err != nil {
		v.AddError(field, "Invalid UUID format", "invalid_uuid")
		return false
	}
	return true
}

// UUIDRequired validates UUID is present and valid
func (v *Validator) UUIDRequired(field, value string) bool {
	if !v.Required(field, value) {
		return false
	}
	return v.UUID(field, value)
}

// ============================================================================
// NUMERIC VALIDATORS
// ============================================================================

// PositiveFloat checks if float is positive (> 0)
func (v *Validator) PositiveFloat(field string, value float64) bool {
	if value <= 0 {
		v.AddError(field, "Must be a positive number", "not_positive")
		return false
	}
	return true
}

// NonNegativeFloat checks if float is >= 0
func (v *Validator) NonNegativeFloat(field string, value float64) bool {
	if value < 0 {
		v.AddError(field, "Must not be negative", "negative")
		return false
	}
	return true
}

// FloatRange checks if float is within range
func (v *Validator) FloatRange(field string, value, min, max float64) bool {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("Must be between %.2f and %.2f", min, max), "out_of_range")
		return false
	}
	return true
}

// PositiveInt checks if int is positive (> 0)
func (v *Validator) PositiveInt(field string, value int) bool {
	if value <= 0 {
		v.AddError(field, "Must be a positive integer", "not_positive")
		return false
	}
	return true
}

// IntRange checks if int is within range
func (v *Validator) IntRange(field string, value, min, max int) bool {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("Must be between %d and %d", min, max), "out_of_range")
		return false
	}
	return true
}

// ============================================================================
// ENUM VALIDATORS
// ============================================================================

// InEnum checks if value is in allowed list
func (v *Validator) InEnum(field, value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	v.AddError(field, fmt.Sprintf("Must be one of: %v", allowed), "invalid_enum")
	return false
}

// ============================================================================
// PATTERN VALIDATORS
// ============================================================================

// Matches validates string against regex pattern
func (v *Validator) Matches(field, value, pattern, message string) bool {
	matched, err := regexp.MatchString(pattern, value)
	if err != nil || !matched {
		v.AddError(field, message, "pattern_mismatch")
		return false
	}
	return true
}

// AlphaNumeric checks if string is alphanumeric
func (v *Validator) AlphaNumeric(field, value string) bool {
	matched := regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(value)
	if !matched {
		v.AddError(field, "Must contain only letters and numbers", "not_alphanumeric")
		return false
	}
	return true
}

// AlphaNumericUnderscore checks if string is alphanumeric with underscores
func (v *Validator) AlphaNumericUnderscore(field, value string) bool {
	matched := regexp.MustCompile(`^[a-zA-Z0-9_]+$`).MatchString(value)
	if !matched {
		v.AddError(field, "Must contain only letters, numbers, and underscores", "invalid_format")
		return false
	}
	return true
}

// Slug validates URL-friendly slug
func (v *Validator) Slug(field, value string) bool {
	matched := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`).MatchString(value)
	if !matched {
		v.AddError(field, "Must be a valid URL slug (lowercase letters, numbers, and hyphens)", "invalid_slug")
		return false
	}
	return true
}

// Phone validates phone number format
func (v *Validator) Phone(field, value string) bool {
	// Basic international phone format
	matched := regexp.MustCompile(`^\+?[0-9]{8,15}$`).MatchString(strings.ReplaceAll(value, " ", ""))
	if !matched {
		v.AddError(field, "Invalid phone number format", "invalid_phone")
		return false
	}
	return true
}

// Domain validates domain name
func (v *Validator) Domain(field, value string) bool {
	matched := regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`).MatchString(value)
	if !matched {
		v.AddError(field, "Invalid domain format", "invalid_domain")
		return false
	}
	return true
}

// ============================================================================
// COMPARISON VALIDATORS
// ============================================================================

// Equals checks if two strings are equal
func (v *Validator) Equals(field, value, expected, message string) bool {
	if value != expected {
		v.AddError(field, message, "not_equal")
		return false
	}
	return true
}

// NotEquals checks if two strings are not equal
func (v *Validator) NotEquals(field, value, forbidden, message string) bool {
	if value == forbidden {
		v.AddError(field, message, "forbidden_value")
		return false
	}
	return true
}
