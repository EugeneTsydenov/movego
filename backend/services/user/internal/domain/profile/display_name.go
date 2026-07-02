package profile

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const maxDisplayNameLength = 32

var (
	ErrDisplayNameEmpty   = errors.New("profile: display name cannot be empty")
	ErrDisplayNameTooLong = errors.New("profile: display name too long")
)

type DisplayName string

func NewDisplayName(raw string) (DisplayName, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", ErrDisplayNameEmpty
	}
	if utf8.RuneCountInString(trimmed) > maxDisplayNameLength {
		return "", ErrDisplayNameTooLong
	}
	return DisplayName(trimmed), nil
}

func (n DisplayName) String() string {
	return string(n)
}
