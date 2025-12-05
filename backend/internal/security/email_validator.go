package security

import (
	"errors"
	"regexp"
	"strings"
)

// AllowedEmailDomain restricts registration to organizational domain only
// This ensures only authorized organization members can access the system
const AllowedEmailDomain = "@lco.org.ly"

var (
	// ErrInvalidEmailFormat indicates the email format is invalid
	ErrInvalidEmailFormat = errors.New("invalid email format")
	// ErrUnauthorizedDomain indicates the email domain is not allowed
	ErrUnauthorizedDomain = errors.New("only @lco.org.ly email addresses are allowed")
)

// emailRegex validates basic email format
// This regex covers most valid email formats while rejecting obviously invalid ones
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// ValidateEmail checks email format and domain restriction
// Returns nil if email is valid and from the allowed domain
func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return ErrInvalidEmailFormat
	}

	if !emailRegex.MatchString(email) {
		return ErrInvalidEmailFormat
	}

	if !strings.HasSuffix(email, AllowedEmailDomain) {
		return ErrUnauthorizedDomain
	}

	return nil
}

// NormalizeEmail converts email to lowercase and trims whitespace
func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}

// IsOrganizationalEmail checks if email belongs to the organization domain
func IsOrganizationalEmail(email string) bool {
	email = NormalizeEmail(email)
	return strings.HasSuffix(email, AllowedEmailDomain)
}
