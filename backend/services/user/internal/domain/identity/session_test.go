package identity

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSession(t *testing.T) {
	accountID := uuid.New()

	session := NewSession(
		accountID,
		"token hash",
		"Chrome",
		"127.0.0.1",
		time.Hour,
	)

	assert.Equal(t, accountID, session.AccountID())
	assert.Equal(t, "token hash", session.TokenHash())
	assert.Equal(t, "Chrome", session.UserAgent())
	assert.Equal(t, "127.0.0.1", session.ClientIP())
	assert.NoError(t, session.Validate())
}

func TestSessionValidate(t *testing.T) {
	tests := []struct {
		name          string
		prepare       func(*Session)
		expectedError error
	}{
		{
			name:          "valid session",
			prepare:       func(*Session) {},
			expectedError: nil,
		},
		{
			name: "revoked session",
			prepare: func(s *Session) {
				s.Revoke()
			},
			expectedError: ErrSessionRevoked,
		},
		{
			name: "expired session",
			prepare: func(s *Session) {
				*s = *NewSession(
					s.AccountID(),
					s.TokenHash(),
					s.UserAgent(),
					s.ClientIP(),
					-time.Minute,
				)
			},
			expectedError: ErrSessionExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := NewSession(
				uuid.New(),
				"hash",
				"Chrome",
				"127.0.0.1",
				time.Hour,
			)

			tt.prepare(session)

			assert.ErrorIs(t, session.Validate(), tt.expectedError)
		})
	}
}

func TestSessionRevoke(t *testing.T) {
	session := NewSession(
		uuid.New(),
		"hash",
		"Chrome",
		"127.0.0.1",
		time.Hour,
	)

	session.Revoke()

	assert.ErrorIs(t, session.Validate(), ErrSessionRevoked)
	assert.NotNil(t, session.RevokedAt())
}

func TestSessionRevokeIsIdempotent(t *testing.T) {
	session := NewSession(
		uuid.New(),
		"hash",
		"Chrome",
		"127.0.0.1",
		time.Hour,
	)

	session.Revoke()
	first := session.RevokedAt()

	session.Revoke()

	assert.Equal(t, first, session.RevokedAt())
}

func TestGenerateRefreshToken(t *testing.T) {
	rawToken, tokenHash, err := GenerateRefreshToken()

	require.NoError(t, err)

	assert.NotEmpty(t, rawToken)
	assert.NotEmpty(t, tokenHash)
	assert.Equal(t, HashRefreshToken(rawToken), tokenHash)
}

func TestHashRefreshToken(t *testing.T) {
	token := "refresh token"

	hash1 := HashRefreshToken(token)
	hash2 := HashRefreshToken(token)

	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
}

func TestHashRefreshTokenDifferentTokens(t *testing.T) {
	hash1 := HashRefreshToken("token 1")
	hash2 := HashRefreshToken("token 2")

	assert.NotEqual(t, hash1, hash2)
}
