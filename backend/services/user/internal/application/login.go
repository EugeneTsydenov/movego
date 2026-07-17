package application

import (
	"context"
	"errors"
	"time"

	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
)

var ErrInvalidCredentials = errors.New("application: invalid email or password")

const refreshTokenTTL = 30 * 24 * time.Hour

type LoginCommand struct {
	Email     string
	Password  string
	UserAgent string
	ClientIP  string
}

type LoginResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

type LoginUseCase struct {
	credentials    identity.CredentialRepository
	authorizations authorization.Repository
	sessions       identity.SessionRepository
	tokenIssuer    TokenIssuer
}

func NewLoginUseCase(
	credentials identity.CredentialRepository,
	authorizations authorization.Repository,
	sessions identity.SessionRepository,
	tokenIssuer TokenIssuer,
) *LoginUseCase {
	return &LoginUseCase{
		credentials:    credentials,
		authorizations: authorizations,
		sessions:       sessions,
		tokenIssuer:    tokenIssuer,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, cmd LoginCommand) (LoginResult, error) {
	email, err := identity.NewEmail(cmd.Email)
	if err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	credential, err := uc.credentials.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, identity.ErrCredentialNotFound) {
			return LoginResult{}, ErrInvalidCredentials
		}
		return LoginResult{}, err
	}

	if err := credential.VerifyPassword(cmd.Password); err != nil {
		return LoginResult{}, ErrInvalidCredentials
	}

	authz, err := uc.authorizations.FindByAccountID(ctx, credential.AccountID())
	if err != nil {
		return LoginResult{}, err
	}

	claims := TokenClaims{
		AccountID: credential.AccountID(),
		Role:      authz.Role().String(),
	}
	accessToken, expiresAt, err := uc.tokenIssuer.IssueAccessToken(ctx, claims)
	if err != nil {
		return LoginResult{}, err
	}

	rawRefreshToken, refreshTokenHash, err := identity.GenerateRefreshToken()
	if err != nil {
		return LoginResult{}, err
	}

	session := identity.NewSession(credential.AccountID(), refreshTokenHash, cmd.UserAgent, cmd.ClientIP, refreshTokenTTL)
	if err := uc.sessions.Save(ctx, session); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		AccessToken:  accessToken,
		RefreshToken: rawRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}
