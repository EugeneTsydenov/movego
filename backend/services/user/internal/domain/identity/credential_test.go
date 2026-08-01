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

	hash := "hashed password"
	credential := NewCredential(accountID, email, hash)

	require.NotNil(t, credential)
	assert.Equal(t, accountID, credential.AccountID())
	assert.Equal(t, email, credential.Email())
	assert.NotEmpty(t, credential.ID())
	assert.Equal(t, hash, credential.PasswordHash())
	assert.False(t, credential.CreatedAt().IsZero())
	assert.False(t, credential.UpdatedAt().IsZero())
	assert.Equal(t, credential.CreatedAt(), credential.UpdatedAt())
}
