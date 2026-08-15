package application

import (
	"context"
	"errors"
	"log/slog"
	"movego/internal/domain"
	"movego/internal/pkg/secret"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AuthService struct {
	uow            UnitOfWork
	credentialRepo domain.CredentialRepo
	sessionRepo    domain.SessionRepo
	tokenIssuer    TokenIssuer
	refreshTTL     time.Duration

	logger *slog.Logger
}

func NewAuthService(
	uow UnitOfWork,
	credentialRepo domain.CredentialRepo,
	sessionRepo domain.SessionRepo,
	tokenIssuer TokenIssuer,
	refreshTTL time.Duration,
	logger *slog.Logger,
) *AuthService {
	return &AuthService{
		uow:            uow,
		credentialRepo: credentialRepo,
		sessionRepo:    sessionRepo,
		tokenIssuer:    tokenIssuer,
		refreshTTL:     refreshTTL,
		logger:         logger,
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

func (s *AuthService) Refresh(ctx context.Context, in RefreshInput) (RefreshOutput, error) {
	parts := strings.SplitN(in.RefreshToken, ".", 2)
	if len(parts) != 2 {
		return RefreshOutput{}, domain.ErrInvalidCredentials
	}
	sessionIDStr := parts[0]
	secrStr := parts[1]
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return RefreshOutput{}, domain.ErrInvalidCredentials
	}

	oldSession, user, err := s.sessionRepo.FindValid(ctx, sessionID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return RefreshOutput{}, domain.ErrInvalidCredentials
		}

		return RefreshOutput{}, err
	}

	if !secret.ValidateToken(secrStr, oldSession.SecretHash()) {
		return RefreshOutput{}, domain.ErrInvalidCredentials
	}

	secr, err := secret.GenerateToken()
	if err != nil {
		return RefreshOutput{}, err
	}

	var session *domain.Session
	err = s.uow.Do(ctx, func(repos domain.Repos) error {
		err := repos.Sessions().Delete(ctx, oldSession.ID())
		if err != nil {
			return err
		}
		session = domain.NewSession(user.ID(), secret.HashToken(secr), oldSession.UserAgent(), oldSession.ClientIP(), s.refreshTTL)
		if err = repos.Sessions().Save(ctx, session); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return RefreshOutput{}, err
	}

	token, err := s.tokenIssuer.Issue(TokenClaims{
		Sub:   user.ID(),
		Roles: []string{user.Role().String()},
	})
	if err != nil {
		return RefreshOutput{}, err
	}

	return RefreshOutput{
		AccessToken:  token,
		RefreshToken: session.ID().String() + "." + secr,
	}, nil
}

func (s *AuthService) SignOut(ctx context.Context, in SignOutInput) error {
	parts := strings.SplitN(in.RefreshToken, ".", 2)
	if len(parts) != 2 {
		return nil
	}
	sessionIDStr := parts[0]

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		return nil
	}

	err = s.sessionRepo.Delete(ctx, sessionID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to delete session from db during sign out",
			slog.String("session_id", sessionID.String()),
			slog.String("error", err.Error()),
		)
		return nil
	}

	return nil
}
