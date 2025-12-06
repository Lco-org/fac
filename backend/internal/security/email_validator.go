package security

import (
	"errors"
	"regexp"
	"strings"
)

const AllowedEmailDomain = "@lco.org.ly"

var (
	ErrInvalidEmailFormat = errors.New("invalid email format")
	ErrUnauthorizedDomain = errors.New("only @lco.org.ly email addresses are allowed")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

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

func NormalizeEmail(email string) string {
	return strings.TrimSpace(strings.ToLower(email))
}
