package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/movego/services/user/internal/domain/account"
	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
)

func seedCredential(t *testing.T, rawPassword string) (*account.Account, *identity.Credential) {
	acc := account.New()
	email, err := identity.NewEmail("player@example.com")
	require.NoError(t, err)
	cred, err := identity.NewCredential(acc.ID(), email, rawPassword)
	require.NoError(t, err)
	return acc, cred
}

func TestLoginUseCase_InvalidEmail(t *testing.T) {
	credentials := newFakeCredentialRepo()
	sessions := newFakeSessionRepo()
	uc := NewLoginUseCase(credentials, newFakeAuthorizationRepo(), sessions, &fakeTokenIssuer{})

	_, err := uc.Execute(context.Background(), LoginCommand{
		Email:    "invalid",
		Password: "password-",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Equal(t, 0, credentials.findByEmailCalls)
	assert.Empty(t, sessions.saved)
}

func TestLoginUseCase_WrongPassword(t *testing.T) {
	ctx := context.Background()
	acc, cred := seedCredential(t, "correct password")
	credentials := newFakeCredentialRepo()
	require.NoError(t, credentials.Save(ctx, cred))

	authzRepo := newFakeAuthorizationRepo()
	require.NoError(t, authzRepo.Save(ctx, authorization.New(acc.ID())))
	sessions := newFakeSessionRepo()
	uc := NewLoginUseCase(credentials, authzRepo, sessions, &fakeTokenIssuer{})

	_, err := uc.Execute(ctx, LoginCommand{
		Email:    "player@example.com",
		Password: "wrong password",
	})

	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Empty(t, sessions.saved)
}

func TestLoginUseCase_RepositoryError(t *testing.T) {
	credentials := newFakeCredentialRepo()
	credentials.findByEmailErr = errors.New("db temporary failure")
	sessions := newFakeSessionRepo()
	uc := NewLoginUseCase(credentials, newFakeAuthorizationRepo(), sessions, &fakeTokenIssuer{})

	_, err := uc.Execute(context.Background(), LoginCommand{
		Email:    "player@example.com",
		Password: "irrelevant",
	})

	require.Error(t, err)
	assert.EqualError(t, err, "db temporary failure")
	assert.Empty(t, sessions.saved)
}

func TestLoginUseCase_Success_IssuesTokenAndPersistsSessionHashOnly(t *testing.T) {
	ctx := context.Background()
	acc, cred := seedCredential(t, "strongpassword123")

	credentials := newFakeCredentialRepo()
	require.NoError(t, credentials.Save(ctx, cred))

	authz := authorization.New(acc.ID())
	authz.PromoteToAdmin()
	authzRepo := newFakeAuthorizationRepo()
	require.NoError(t, authzRepo.Save(ctx, authz))

	sessions := newFakeSessionRepo()
	expectedExpiry := time.Now().UTC().Add(30 * time.Minute).Round(time.Second)
	issuer := &fakeTokenIssuer{
		accessToken: "access token",
		expiresAt:   expectedExpiry,
	}
	uc := NewLoginUseCase(credentials, authzRepo, sessions, issuer)

	result, err := uc.Execute(ctx, LoginCommand{
		Email:     "player@example.com",
		Password:  "strongpassword123",
		UserAgent: "android",
		ClientIP:  "127.0.0.1",
	})
	require.NoError(t, err)
	require.Len(t, sessions.saved, 1)

	assert.Equal(t, "access token", result.AccessToken)
	assert.Equal(t, expectedExpiry, result.ExpiresAt)
	assert.NotEmpty(t, result.RefreshToken)
	assert.Equal(t, acc.ID(), issuer.lastClaims.AccountID)
	assert.Equal(t, authorization.RoleAdmin.String(), issuer.lastClaims.Role)

	var persisted *identity.Session
	for _, s := range sessions.saved {
		persisted = s
	}
	require.NotNil(t, persisted)
	assert.Equal(t, acc.ID(), persisted.AccountID())
	assert.Equal(t, "android", persisted.UserAgent())
	assert.Equal(t, "127.0.0.1", persisted.ClientIP())
	assert.Equal(t, identity.HashRefreshToken(result.RefreshToken), persisted.TokenHash())
	assert.NotEqual(t, result.RefreshToken, persisted.TokenHash(), "raw token must never be stored")
}
