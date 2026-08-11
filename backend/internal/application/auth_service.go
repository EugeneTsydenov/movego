package application

import (
	"context"
	"movego/internal/domain"
	"movego/internal/pkg/secret"
	"time"
)

type AuthService struct {
	uow         UnitOfWork
	tokenIssuer TokenIssuer
	refreshTTL  time.Duration
}

func NewAuthService(uow UnitOfWork, tokenIssuer TokenIssuer, refreshTTL time.Duration) *AuthService {
	return &AuthService{
		uow:         uow,
		tokenIssuer: tokenIssuer,
		refreshTTL:  refreshTTL,
	}
}

func (s *AuthService) SignUp(ctx context.Context, in SignUpInput) (SignUpOutput, error) {
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return SignUpOutput{}, err
	}

	password, err := domain.NewPlainPassword(in.Password)
	if err != nil {
		return SignUpOutput{}, err
	}

	hash, err := secret.HashPassword(password.String())
	if err != nil {
		return SignUpOutput{}, err
	}

	user := domain.NewTemporaryUser(email, domain.RoleUser)
	var session *domain.Session
	var secr string
	err = s.uow.Do(ctx, func(repos domain.Repos) error {
		err = repos.Users().Save(ctx, user)
		if err != nil {
			return err
		}

		credential := domain.NewPasswordCredential(user.ID(), domain.Password, hash)
		err = repos.Credentials().Save(ctx, credential)
		if err != nil {
			return err
		}

		secr, err = secret.GenerateToken()
		if err != nil {
			return err
		}

		session = domain.NewSession(user.ID(), secret.HashToken(secr), in.UserAgent, in.ClientIP, s.refreshTTL)
		err = repos.Sessions().Save(ctx, session)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return SignUpOutput{}, err
	}

	token, err := s.tokenIssuer.Issue(TokenClaims{
		Sub:   user.ID(),
		Roles: []string{domain.RoleUser.String()},
	})
	if err != nil {
		return SignUpOutput{}, err
	}

	return SignUpOutput{
		User: UserDTO{
			ID:          user.ID(),
			Tag:         user.Tag().String(),
			Email:       email.String(),
			DisplayName: user.DisplayName().String(),
			Role:        domain.RoleUser.String(),
			UpdatedAt:   user.UpdatedAt(),
			CreatedAt:   user.CreatedAt(),
		},
		AccessToken:  token,
		RefreshToken: session.ID().String() + "." + secr,
	}, nil
}

func (s *AuthService) SignIn(ctx context.Context) {

}
