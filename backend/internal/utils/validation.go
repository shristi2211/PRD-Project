package utils

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

// ValidateRegisterInput validates registration fields and returns a human-readable error.
func ValidateRegisterInput(email, password, name string) error {
	email = strings.TrimSpace(email)
	name = strings.TrimSpace(name)

	if email == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return fmt.Errorf("invalid email format")
	}

	if password == "" {
		return fmt.Errorf("password is required")
	}
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}
	if len(password) > 72 {
		// bcrypt has a 72-byte limit
		return fmt.Errorf("password must not exceed 72 characters")
	}
	if !hasMinComplexity(password) {
		return fmt.Errorf("password must contain at least one uppercase letter, one lowercase letter, and one digit")
	}

	if name == "" {
		return fmt.Errorf("name is required")
	}
	if len(name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}
	if len(name) > 255 {
		return fmt.Errorf("name must not exceed 255 characters")
	}

	return nil
}

// ValidateLoginInput validates login fields.
func ValidateLoginInput(email, password string) error {
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("email is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func hasMinComplexity(password string) bool {
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		}
	}
	return hasUpper && hasLower && hasDigit
}
