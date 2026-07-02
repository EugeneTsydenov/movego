package profile

import (
	"errors"
	"strings"
)

var ErrInvalidAvatarURL = errors.New("profile: avatar url must start with http:// or https://")

type AvatarURL string

func NewAvatarURL(raw string) (AvatarURL, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", nil
	}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return "", ErrInvalidAvatarURL
	}
	return AvatarURL(trimmed), nil
}

func (u AvatarURL) IsSet() bool {
	return u != ""
}

func (u AvatarURL) String() string {
	return string(u)
}
