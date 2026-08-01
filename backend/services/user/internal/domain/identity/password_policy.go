package identity

import "github.com/movego/services/user/internal/domain"

const minLen = 8
const maxLen = 128

func ValidatePassword(raw string) error {
	if len(raw) < minLen && len(raw) > 128 {
		return domain.NewInvalidInput("credential.password_too_weak", "password too weak")
	}

	return nil
}
