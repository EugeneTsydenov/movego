package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/movego/services/user/internal/domain/account"
	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
)

var (
	ErrAccountInactive     = errors.New("usecase: account is inactive")
	ErrInvalidRefreshToken = errors.New("usecase: invalid or expired refresh token")
)

type RefreshTokenCommand struct {
	RefreshToken string
	UserAgent    string
	ClientIP     string
}

type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type RefreshTokenUseCase struct {
	accounts       account.Repository
	authorizations authorization.Repository
	sessions       identity.SessionRepository
	uow            UOW
	tokenIssuer    TokenIssuer
}

func NewRefreshTokenUseCase(
	accounts account.Repository,
	authorizations authorization.Repository,
	sessions identity.SessionRepository,
	uow UOW,
	tokenIssuer TokenIssuer,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		accounts:       accounts,
		authorizations: authorizations,
		sessions:       sessions,
		uow:            uow,
		tokenIssuer:    tokenIssuer,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, cmd RefreshTokenCommand) (RefreshTokenResult, error) {
	tokenHash := identity.HashRefreshToken(cmd.RefreshToken)

	session, err := uc.sessions.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, identity.ErrSessionNotFound) {
			return RefreshTokenResult{}, ErrInvalidRefreshToken
		}
		return RefreshTokenResult{}, err
	}

	if err := session.Validate(); err != nil {
		return RefreshTokenResult{}, ErrInvalidRefreshToken
	}

	acc, err := uc.accounts.FindByID(ctx, session.AccountID())
	if err != nil {
		return RefreshTokenResult{}, err
	}
	if !acc.IsActive() {
		return RefreshTokenResult{}, ErrAccountInactive
	}

	authz, err := uc.authorizations.FindByAccountID(ctx, session.AccountID())
	if err != nil {
		return RefreshTokenResult{}, err
	}

	accessToken, expiresAt, err := uc.tokenIssuer.IssueAccessToken(ctx, TokenClaims{
		AccountID: session.AccountID(),
		Role:      authz.Role().String(),
	})
	if err != nil {
		return RefreshTokenResult{}, err
	}

	newRawToken, newTokenHash, err := identity.GenerateRefreshToken()
	if err != nil {
		return RefreshTokenResult{}, err
	}

	session.Revoke()
	newSession := identity.NewSession(session.AccountID(), newTokenHash, cmd.UserAgent, cmd.ClientIP, refreshTokenTTL)

	err = uc.uow.Do(ctx, func(repos *Repositories) error {
		if err := repos.Sessions.Save(ctx, session); err != nil {
			return err
		}
		if err := repos.Sessions.Save(ctx, newSession); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return RefreshTokenResult{}, err
	}

	return RefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: newRawToken,
		ExpiresAt:    expiresAt,
	}, nil
}
