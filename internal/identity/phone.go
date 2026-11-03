package identity

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrEmptyPhone    = errors.New("phone number cannot be empty")
	ErrTooShort      = errors.New("phone number too short (minimum 7 digits)")
	ErrTooLong       = errors.New("phone number too long (maximum 15 digits per ITU E.164)")
	ErrInvalidFormat = errors.New("invalid phone number format")
	ErrNonDigitChars = errors.New("phone number contains invalid non-numeric characters")
)

// NormalizeE164 normalizes an incoming phone number into standard E.164 format: +<country_code><subscriber_number>.
// It strips whitespace, dashes, parentheses, dots, and ensures standard formatting.
func NormalizeE164(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrEmptyPhone
	}

	var digits strings.Builder
	hasLeadingPlus := false

	for i, r := range trimmed {
		if i == 0 && r == '+' {
			hasLeadingPlus = true
			continue
		}
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '(' || r == ')' || r == '.' {
			// standard formatting characters, skip
			continue
		} else {
			return "", fmt.Errorf("%w: '%c'", ErrNonDigitChars, r)
		}
	}

	digitStr := digits.String()
	if len(digitStr) < 7 {
		return "", ErrTooShort
	}
	if len(digitStr) > 15 {
		return "", ErrTooLong
	}

	// Always prefix with +
	_ = hasLeadingPlus
	return "+" + digitStr, nil
}

// CompareE164 securely checks whether two phone numbers represent the same normalized E.164 number.
// If either number fails to normalize, it fails closed (returns false).
func CompareE164(a, b string) bool {
	normA, errA := NormalizeE164(a)
	if errA != nil {
		return false
	}
	normB, errB := NormalizeE164(b)
	if errB != nil {
		return false
	}
	return normA == normB
}

// MaskPhone masks the central digits of a phone number to protect operator privacy in logs, emails, and UI.
// Example: "+19876543210" -> "+1987•••••10"
func MaskPhone(phone string) string {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return "Unconfigured"
	}
	clean := strings.ReplaceAll(trimmed, " ", "")
	if len(clean) <= 6 {
		return strings.Repeat("•", len(clean))
	}
	// Preserve country prefix (e.g. +91 95) and last 2 digits
	prefixLen := 5
	if len(clean) < 10 {
		prefixLen = 3
	}
	prefix := clean[:prefixLen]
	suffix := clean[len(clean)-2:]
	maskLen := len(clean) - len(prefix) - len(suffix)
	if maskLen < 3 {
		maskLen = 4
	}
	return prefix + strings.Repeat("•", maskLen) + suffix
}

// MaskEmail masks the username portion of an email address to prevent privacy leaks.
// Example: "operator.bridge@example.com" -> "op••••ge@example.com"
func MaskEmail(email string) string {
	trimmed := strings.TrimSpace(email)
	if trimmed == "" {
		return "Unconfigured"
	}
	parts := strings.Split(trimmed, "@")
	if len(parts) != 2 {
		return "••••@" + trimmed
	}
	user := parts[0]
	domain := parts[1]
	if len(user) <= 4 {
		return string(user[0]) + "••@" + domain
	}
	prefix := user[:2]
	suffix := user[len(user)-2:]
	return prefix + "••••" + suffix + "@" + domain
}
