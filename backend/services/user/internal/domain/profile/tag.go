package profile

import (
	"errors"
	"regexp"
	"strings"
)

var (
	tagPattern    = regexp.MustCompile(`^[a-z0-9_]{3,20}$`)
	ErrInvalidTag = errors.New("profile: tag must be 3-20 chars, lowercase letters/digits/underscore")
)

type Tag string

func NewTag(raw string) (Tag, error) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	if !tagPattern.MatchString(normalized) {
		return "", ErrInvalidTag
	}
	return Tag(normalized), nil
}

func (t Tag) String() string {
	return string(t)
}
