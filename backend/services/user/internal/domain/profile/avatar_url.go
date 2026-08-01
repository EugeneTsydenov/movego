package profile

import (
	"strings"

	"github.com/movego/services/user/internal/domain"
)

type AvatarURL string

func NewAvatarURL(raw string) (AvatarURL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return "", domain.NewInvalidInput("profile.invalid_avatar_url", "invalid avatar url")
	}
	return AvatarURL(trimmed), nil
}

func (u AvatarURL) IsSet() bool {
	return u != ""
}

func (u AvatarURL) String() string {
	return string(u)
}
