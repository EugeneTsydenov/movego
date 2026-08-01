package identity

import (
	"net/mail"
	"strings"

	"github.com/movego/services/user/internal/domain"
)

// Value object
type Email string

func (e Email) String() string {
	return string(e)
}

func NewEmail(v string) (Email, error) {
	trimmed := strings.TrimSpace(v)
	addr, err := mail.ParseAddress(trimmed)
	if err != nil {
		return "", domain.NewInvalidInput("credential.invalid_email", "invalid email")
	}

	return Email(strings.ToLower(addr.Address)), nil
}
