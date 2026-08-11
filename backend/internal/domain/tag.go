package domain

import (
	"fmt"
	"regexp"
)

const (
	minTagLength = 3
	maxTagLength = 20
)

var tagPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

type Tag struct {
	value string
}

func NewTag(tagStr string) (Tag, error) {
	if len(tagStr) < minTagLength || len(tagStr) > maxTagLength {
		return Tag{}, fmt.Errorf("%w: %q", ErrInvalidTag, tagStr)
	}
	for _, r := range tagStr {
		isLetter := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		isDigit := r >= '0' && r <= '9'
		if !isLetter && !isDigit {
			return Tag{}, fmt.Errorf("%w: %q", ErrInvalidTag, tagStr)
		}
	}
	return Tag{
		value: tagStr,
	}, nil
}

func (t Tag) String() string {
	return t.value
}
