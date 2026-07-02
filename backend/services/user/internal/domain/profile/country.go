package profile

import (
	"errors"
	"strings"
)

var ErrInvalidCountryCode = errors.New("profile: country code must be a 2-letter ISO code")

type CountryCode string

func newCountryCode(raw string) (CountryCode, error) {
	trimmed := strings.ToUpper(strings.TrimSpace(raw))
	if trimmed == "" {
		return "", nil
	}
	if len(trimmed) != 2 {
		return "", ErrInvalidCountryCode
	}
	for _, r := range trimmed {
		if r < 'A' || r > 'Z' {
			return "", ErrInvalidCountryCode
		}
	}
	return CountryCode(trimmed), nil
}

func (c CountryCode) IsSet() bool {
	return c != ""
}

func (c CountryCode) String() string {
	return string(c)
}
