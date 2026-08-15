package application

import (
	"context"
	"errors"
	"movego/internal/domain"
	"movego/internal/pkg/secret"
	"time"
)

type AuthService struct {
	uow            UnitOfWork
	credentialRepo domain.CredentialRepo
	sessionRepo    domain.SessionRepo
	tokenIssuer    TokenIssuer
	refreshTTL     time.Duration
}

func NewAuthService(
	uow UnitOfWork,
	credentialRepo domain.CredentialRepo,
	sessionRepo domain.SessionRepo,
	tokenIssuer TokenIssuer,
	refreshTTL time.Duration,
) *AuthService {
	return &AuthService{
		uow:            uow,
		credentialRepo: credentialRepo,
		sessionRepo:    sessionRepo,
		tokenIssuer:    tokenIssuer,
		refreshTTL:     refreshTTL,
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

	secr, err := secret.GenerateToken()
	if err != nil {
		return SignUpOutput{}, err
	}

	user := domain.NewTemporaryUser(email, domain.RoleUser)
	var session *domain.Session
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

func (s *AuthService) SignIn(ctx context.Context, in SignInInput) (SignInOutput, error) {
	email, err := domain.NewEmail(in.Email)
	if err != nil {
		return SignInOutput{}, err
	}

	password, err := domain.NewPlainPassword(in.Password)
	if err != nil {
		return SignInOutput{}, err
	}

	user, credential, err := s.credentialRepo.FindForAuth(ctx, email, domain.Password)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return SignInOutput{}, domain.ErrInvalidCredentials
		}

		return SignInOutput{}, err
	}

	hash := credential.PasswordHash()
	if hash == nil || !secret.CheckPasswordHash(password.String(), *hash) {
		return SignInOutput{}, domain.ErrInvalidCredentials
	}

	secr, err := secret.GenerateToken()
	if err != nil {
		return SignInOutput{}, err
	}

	session := domain.NewSession(user.ID(), secret.HashToken(secr), in.UserAgent, in.ClientIP, s.refreshTTL)
	err = s.sessionRepo.Save(ctx, session)
	if err != nil {
		return SignInOutput{}, err
	}

	token, err := s.tokenIssuer.Issue(TokenClaims{
		Sub:   user.ID(),
		Roles: []string{user.Role().String()},
	})
	if err != nil {
		return SignInOutput{}, err
	}

	return SignInOutput{
		User: UserDTO{
			ID:          user.ID(),
			Tag:         user.Tag().String(),
			Email:       email.String(),
			DisplayName: user.DisplayName().String(),
			Role:        user.Role().String(),
			UpdatedAt:   user.UpdatedAt(),
			CreatedAt:   user.CreatedAt(),
		},
		AccessToken:  token,
		RefreshToken: session.ID().String() + "." + secr,
	}, nil
}
