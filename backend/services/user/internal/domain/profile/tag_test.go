package profile

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected Tag
		wantErr  error
	}{
		{
			name:     "valid tag",
			input:    "player123",
			expected: Tag("player123"),
		},
		{
			name:     "normalizes uppercase",
			input:    "Player123",
			expected: Tag("player123"),
		},
		{
			name:     "trims spaces",
			input:    "  player123  ",
			expected: Tag("player123"),
		},
		{
			name:     "allows underscore",
			input:    "pro_player",
			expected: Tag("pro_player"),
		},
		{
			name:    "too short",
			input:   "ab",
			wantErr: ErrInvalidTag,
		},
		{
			name:    "too long",
			input:   "abcdefghijklmnopqrstu",
			wantErr: ErrInvalidTag,
		},
		{
			name:    "contains spaces",
			input:   "player 123",
			wantErr: ErrInvalidTag,
		},
		{
			name:    "contains dash",
			input:   "player-123",
			wantErr: ErrInvalidTag,
		},
		{
			name:    "contains special characters",
			input:   "player!",
			wantErr: ErrInvalidTag,
		},
		{
			name:    "empty",
			input:   "",
			wantErr: ErrInvalidTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag, err := NewTag(tt.input)

			assert.ErrorIs(t, err, tt.wantErr)

			if tt.wantErr == nil {
				assert.Equal(t, tt.expected, tag)
				assert.Equal(t, string(tt.expected), tag.String())
			}
		})
	}
}

func TestTagString(t *testing.T) {
	tag := Tag("player123")

	assert.Equal(t, "player123", tag.String())
}
