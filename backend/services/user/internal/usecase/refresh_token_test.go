package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/movego/services/user/internal/domain/account"
	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
)

func TestRefreshTokenUseCase_UnknownToken_ReturnsInvalidRefreshToken(t *testing.T) {
	ctx := context.Background()
	repos := newFakeRepositories()
	uc := NewRefreshTokenUseCase(
		repos.Accounts,
		repos.Authorizations,
		repos.Sessions,
		&fakeUOW{repos: repos},
		&fakeTokenIssuer{},
	)

	_, err := uc.Execute(ctx, RefreshTokenCommand{
		RefreshToken: "refresh token",
	})

	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func TestRefreshTokenUseCase_RevokedSessionToken_ReturnsInvalidRefreshToken(t *testing.T) {
	ctx := context.Background()
	acc := account.New()
	raw, hash, err := identity.GenerateRefreshToken()
	require.NoError(t, err)

	repos := newFakeRepositories()
	require.NoError(t, repos.Accounts.Save(ctx, acc))
	require.NoError(t, repos.Authorizations.Save(ctx, authorization.New(acc.ID())))

	session := identity.NewSession(acc.ID(), hash, "android", "127.0.0.1", time.Hour)
	session.Revoke()
	require.NoError(t, repos.Sessions.Save(ctx, session))

	uc := NewRefreshTokenUseCase(
		repos.Accounts,
		repos.Authorizations,
		repos.Sessions,
		&fakeUOW{repos: repos},
		&fakeTokenIssuer{},
	)

	_, err = uc.Execute(ctx, RefreshTokenCommand{
		RefreshToken: raw,
		UserAgent:    "android",
		ClientIP:     "127.0.0.1",
	})

	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func TestRefreshTokenUseCase_InactiveAccount_ReturnsAccountInactive(t *testing.T) {
	ctx := context.Background()
	acc := account.New()
	require.NoError(t, acc.Suspend())
	raw, hash, err := identity.GenerateRefreshToken()
	require.NoError(t, err)

	repos := newFakeRepositories()
	require.NoError(t, repos.Accounts.Save(ctx, acc))
	require.NoError(t, repos.Authorizations.Save(ctx, authorization.New(acc.ID())))
	session := identity.NewSession(acc.ID(), hash, "android", "127.0.0.1", time.Hour)
	require.NoError(t, repos.Sessions.Save(ctx, session))

	uc := NewRefreshTokenUseCase(
		repos.Accounts,
		repos.Authorizations,
		repos.Sessions,
		&fakeUOW{repos: repos},
		&fakeTokenIssuer{},
	)

	_, err = uc.Execute(ctx, RefreshTokenCommand{
		RefreshToken: raw,
		UserAgent:    "android",
		ClientIP:     "127.0.0.1",
	})

	assert.ErrorIs(t, err, ErrAccountInactive)
}

func TestRefreshTokenUseCase_Success_RotatesTokenAndRevokesOldSession(t *testing.T) {
	ctx := context.Background()
	acc := account.New()
	oldRaw, oldHash, err := identity.GenerateRefreshToken()
	require.NoError(t, err)

	oldSession := identity.NewSession(acc.ID(), oldHash, "android", "127.0.0.1", time.Hour)

	repos := newFakeRepositories()
	require.NoError(t, repos.Accounts.Save(ctx, acc))
	adminAuthz := authorization.New(acc.ID())
	adminAuthz.PromoteToAdmin()
	require.NoError(t, repos.Authorizations.Save(ctx, adminAuthz))
	require.NoError(t, repos.Sessions.Save(ctx, oldSession))

	exp := time.Now().UTC().Add(20 * time.Minute).Round(time.Second)
	issuer := &fakeTokenIssuer{accessToken: "new access token", expiresAt: exp}
	uc := NewRefreshTokenUseCase(
		repos.Accounts,
		repos.Authorizations,
		repos.Sessions,
		&fakeUOW{repos: repos},
		issuer,
	)
	saveCallsBeforeRotate := repos.Sessions.(*fakeSessionRepo).saveCalls

	result, err := uc.Execute(ctx, RefreshTokenCommand{
		RefreshToken: oldRaw,
		UserAgent:    "android",
		ClientIP:     "127.0.0.1",
	})
	require.NoError(t, err)

	assert.Equal(t, "new access token", result.AccessToken)
	assert.Equal(t, exp, result.ExpiresAt)
	assert.NotEmpty(t, result.RefreshToken)
	assert.True(t, oldSession.IsRevoked(), "old refresh session must be revoked on rotation")
	assert.Equal(t, saveCallsBeforeRotate+2, repos.Sessions.(*fakeSessionRepo).saveCalls, "rotation must persist old revoked and new session")

	newSession, err := repos.Sessions.(*fakeSessionRepo).findNonRevokedByTokenHash(identity.HashRefreshToken(result.RefreshToken))
	require.NoError(t, err)
	assert.Equal(t, acc.ID(), newSession.AccountID())
	assert.Equal(t, "127.0.0.1", newSession.ClientIP())

	_, err = uc.Execute(ctx, RefreshTokenCommand{
		RefreshToken: oldRaw,
		UserAgent:    "android",
		ClientIP:     "127.0.0.1",
	})
	assert.ErrorIs(t, err, ErrInvalidRefreshToken, "old refresh token must be single-use after rotation")
}
