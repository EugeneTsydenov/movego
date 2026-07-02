package identity

import (
	"errors"
	"net/mail"
	"strings"
)

var (
	ErrInvalidEmail = errors.New("identity: invalid email format")
)

type Email string

func (e Email) String() string {
	return string(e)
}

func NewEmail(v string) (Email, error) {
	trimmed := strings.TrimSpace(v)
	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", ErrInvalidEmail
	}

	return Email(strings.ToLower(addr.Address)), nil
}
