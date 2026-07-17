package identity

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCredential(t *testing.T) {
	accountID := uuid.New()
	email, err := NewEmail("test@example.com")
	require.NoError(t, err)

	credential, err := NewCredential(accountID, email, "password123")

	require.NoError(t, err)
	require.NotNil(t, credential)

	assert.Equal(t, accountID, credential.AccountID())
	assert.Equal(t, email.String(), credential.Email())
	assert.NotEmpty(t, credential.ID())
	assert.NotEmpty(t, credential.PasswordHash())
	assert.NotEqual(t, "password123", credential.PasswordHash())
	assert.False(t, credential.CreatedAt().IsZero())
	assert.False(t, credential.UpdatedAt().IsZero())
	assert.Equal(t, credential.CreatedAt(), credential.UpdatedAt())

	assert.NoError(t, credential.VerifyPassword("password123"))
}

func TestNewCredentialWeakPassword(t *testing.T) {
	accountID := uuid.New()
	email, err := NewEmail("test@example.com")
	require.NoError(t, err)

	credential, err := NewCredential(accountID, email, "1234567")

	assert.Nil(t, credential)
	assert.ErrorIs(t, err, ErrWeakPassword)
}

func TestVerifyPassword(t *testing.T) {
	accountID := uuid.New()
	email, err := NewEmail("test@example.com")
	require.NoError(t, err)

	credential, err := NewCredential(accountID, email, "password123")
	require.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		expectedErr error
	}{
		{
			name:        "correct password",
			password:    "password123",
			expectedErr: nil,
		},
		{
			name:        "incorrect password",
			password:    "wrong-password",
			expectedErr: ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := credential.VerifyPassword(tt.password)
			assert.ErrorIs(t, err, tt.expectedErr)
		})
	}
}

func TestChangePassword(t *testing.T) {
	accountID := uuid.New()
	email, err := NewEmail("test@example.com")
	require.NoError(t, err)

	credential, err := NewCredential(accountID, email, "oldpassword")
	require.NoError(t, err)

	oldHash := credential.PasswordHash()

	err = credential.ChangePassword("newpassword")

	require.NoError(t, err)

	assert.NotEqual(t, oldHash, credential.PasswordHash())
	assert.True(t, credential.UpdatedAt().After(credential.CreatedAt()))

	assert.NoError(t, credential.VerifyPassword("newpassword"))
	assert.ErrorIs(t, credential.VerifyPassword("oldpassword"), ErrInvalidCredentials)
}

func TestChangePasswordWeakPassword(t *testing.T) {
	accountID := uuid.New()
	email, err := NewEmail("test@example.com")
	require.NoError(t, err)

	credential, err := NewCredential(accountID, email, "password123")
	require.NoError(t, err)

	oldHash := credential.PasswordHash()

	err = credential.ChangePassword("1234567")

	assert.ErrorIs(t, err, ErrWeakPassword)
	assert.Equal(t, oldHash, credential.PasswordHash())
	assert.NoError(t, credential.VerifyPassword("password123"))
}
