package domain

import (
	"fmt"
	"net/mail"
)

type Email struct {
	value string
}

func NewEmail(emailStr string) (Email, error) {
	if _, err := mail.ParseAddress(emailStr); err != nil {
		return Email{}, fmt.Errorf("%w: %q", ErrInvalidEmail, emailStr)
	}
	return Email{
		value: emailStr,
	}, nil
}

func (e Email) String() string {
	return e.value
}
