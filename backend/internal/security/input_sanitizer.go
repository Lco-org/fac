package security

import (
	"regexp"
	"strings"
	"unicode"
)

// SQL injection patterns to detect and block
var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(\b(union|select|insert|update|delete|drop|truncate|alter|exec|execute)\b.*\b(from|into|table|database)\b)`),
	regexp.MustCompile(`(?i)(\b(or|and)\b\s*[\d\w'"]+\s*=\s*[\d\w'"]+)`),
	regexp.MustCompile(`(?i)(--|#|/\*|\*/)`),
	regexp.MustCompile(`(?i)(\bsleep\s*\(|\bbenchmark\s*\(|\bwaitfor\s+delay\b)`),
	regexp.MustCompile(`(?i)(\bchar\s*\(|\bconcat\s*\(|\bconvert\s*\()`),
}

// XSS patterns to detect and sanitize
var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?i)<[^>]*\son\w+\s*=`),
	regexp.MustCompile(`(?i)javascript:`),
	regexp.MustCompile(`(?i)vbscript:`),
	regexp.MustCompile(`(?i)data:text/html`),
}

// InputSanitizer provides methods for sanitizing and validating user input
type InputSanitizer struct{}

// NewInputSanitizer creates a new InputSanitizer instance
func NewInputSanitizer() *InputSanitizer {
	return &InputSanitizer{}
}

// ContainsSQLInjection checks if input contains potential SQL injection patterns
func (s *InputSanitizer) ContainsSQLInjection(input string) bool {
	normalizedInput := strings.ToLower(strings.TrimSpace(input))
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(normalizedInput) {
			return true
		}
	}
	return false
}

// ContainsXSS checks if input contains potential XSS patterns
func (s *InputSanitizer) ContainsXSS(input string) bool {
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}
	return false
}

// SanitizeString removes potentially dangerous characters and patterns
func (s *InputSanitizer) SanitizeString(input string) string {
	// Remove null bytes
	result := strings.ReplaceAll(input, "\x00", "")

	// Remove common XSS vectors
	for _, pattern := range xssPatterns {
		result = pattern.ReplaceAllString(result, "")
	}

	// Escape HTML entities
	result = strings.ReplaceAll(result, "<", "&lt;")
	result = strings.ReplaceAll(result, ">", "&gt;")
	result = strings.ReplaceAll(result, "\"", "&quot;")
	result = strings.ReplaceAll(result, "'", "&#x27;")

	return result
}

// ValidateFinancialAmount validates that a string represents a valid financial amount
func (s *InputSanitizer) ValidateFinancialAmount(amount float64) bool {
	// Amount must be positive and have at most 2 decimal places
	if amount <= 0 {
		return false
	}
	// Check for reasonable maximum (10 billion)
	if amount > 10_000_000_000 {
		return false
	}
	return true
}

// SanitizeUUID validates and returns a clean UUID string or empty if invalid
func (s *InputSanitizer) SanitizeUUID(input string) string {
	// UUID pattern: 8-4-4-4-12 hex characters
	uuidPattern := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	input = strings.TrimSpace(input)
	if uuidPattern.MatchString(input) {
		return strings.ToLower(input)
	}
	return ""
}

// SanitizeAlphanumeric removes all non-alphanumeric characters except spaces and common punctuation
func (s *InputSanitizer) SanitizeAlphanumeric(input string) string {
	var result strings.Builder
	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) || r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// ValidateDescription ensures description length and content are acceptable
func (s *InputSanitizer) ValidateDescription(input string, maxLength int) (string, bool) {
	input = strings.TrimSpace(input)
	if len(input) > maxLength {
		return "", false
	}
	if s.ContainsSQLInjection(input) || s.ContainsXSS(input) {
		return "", false
	}
	return input, true
}

// ValidateReferenceNumber validates transaction reference numbers
func (s *InputSanitizer) ValidateReferenceNumber(input string) (string, bool) {
	input = strings.TrimSpace(input)
	// Reference numbers should be alphanumeric with optional dashes
	refPattern := regexp.MustCompile(`^[A-Za-z0-9\-_]{1,50}$`)
	if refPattern.MatchString(input) {
		return strings.ToUpper(input), true
	}
	return "", false
}
