package domain

import (
	"fmt"
	"unicode/utf8"
)

const (
	minDisplayNameLength = 3
	maxDisplayNameLength = 12
)

type DisplayName struct {
	value string
}

func NewDisplayName(displayNameStr string) (DisplayName, error) {
	length := utf8.RuneCountInString(displayNameStr)
	if length < minDisplayNameLength || length > maxDisplayNameLength {
		return DisplayName{}, fmt.Errorf("%w: %q", ErrInvalidDisplayName, displayNameStr)
	}
	return DisplayName{
		value: displayNameStr,
	}, nil
}

func (d DisplayName) String() string {
	return d.value
}
