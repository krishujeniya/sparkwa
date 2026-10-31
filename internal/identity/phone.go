package identity

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrEmptyPhone     = errors.New("phone number cannot be empty")
	ErrTooShort       = errors.New("phone number too short (minimum 7 digits)")
	ErrTooLong        = errors.New("phone number too long (maximum 15 digits per ITU E.164)")
	ErrInvalidFormat  = errors.New("invalid phone number format")
	ErrNonDigitChars  = errors.New("phone number contains invalid non-numeric characters")
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
