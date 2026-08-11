package domain

import (
	"unicode"
	"unicode/utf8"
)

const minPasswordLength = 8
const maxPasswordLength = 72

type PlainPassword struct {
	value string
}

func NewPlainPassword(passwordStr string) (PlainPassword, error) {
	if utf8.RuneCountInString(passwordStr) < minPasswordLength || utf8.RuneCountInString(passwordStr) > maxPasswordLength {
		return PlainPassword{}, ErrWeakPassword
	}

	var hasUpper, hasLower, hasDigit bool
	for _, r := range passwordStr {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit {
		return PlainPassword{}, ErrWeakPassword
	}

	return PlainPassword{
		value: passwordStr,
	}, nil
}

func (p PlainPassword) String() string {
	return p.value
}
