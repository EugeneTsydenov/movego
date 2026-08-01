package profile

import (
	"strings"

	"github.com/movego/services/user/internal/domain"
)

var ErrInvalidCountryCode = domain.NewInvalidInput("profile.invalid_country_code", "invalid country code")

type CountryCode string

func NewCountryCode(raw string) (CountryCode, error) {
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
