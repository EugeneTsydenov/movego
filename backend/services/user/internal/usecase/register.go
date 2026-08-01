package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/movego/services/user/internal/domain/account"
	"github.com/movego/services/user/internal/domain/authorization"
	"github.com/movego/services/user/internal/domain/identity"
	"github.com/movego/services/user/internal/domain/profile"
	"github.com/movego/shared/validation"
)

var (
	ErrEmailAlreadyTaken   = errors.New("email already taken")
	ErrTagGenerationFailed = errors.New("failed to generate unique tag after multiple attempts")
)

type RegisterCommand struct {
	Email       string
	Password    string
	DisplayName string
}

type RegisterUseCase struct {
	uow            UOW
	passwordHasher PasswordHasher
}

func NewRegisterUseCase(uow UOW, passwordHasher PasswordHasher) *RegisterUseCase {
	return &RegisterUseCase{uow, passwordHasher}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, cmd RegisterCommand) error {
	valErrs := validation.New()

	email, err := identity.NewEmail(cmd.Email)
	if err != nil {
		valErrs.Add(err)
	}

	err = identity.ValidatePassword(cmd.Password)
	if err != nil {
		valErrs.Add(err)
	}

	displayName, err := profile.NewDisplayName(cmd.DisplayName)
	if err != nil {
		valErrs.Add(err)
	}

	if valErrs.HasErrors() {
		return valErrs
	}

	return uc.uow.Do(ctx, func(repos *Repositories) error {
		exist, err := repos.Credentials.ExistsByEmail(ctx, email)
		if err != nil {
			return err
		}

		if exist {
			return ErrEmailAlreadyTaken
		}

		acc := account.New()

		hashPassword, err := uc.passwordHasher.Hash(cmd.Password)
		if err != nil {
			return err
		}

		credential := identity.NewCredential(acc.ID(), email, hashPassword)
		if err != nil {
			return err
		}

		authz := authorization.New(acc.ID())

		tag, err := uc.generateUniqueTag(ctx, repos)
		if err != nil {
			return err
		}

		prof := profile.New(acc.ID(), tag, displayName, "", "")

		if err = repos.Accounts.Save(ctx, acc); err != nil {
			return err
		}

		if err = repos.Credentials.Save(ctx, credential); err != nil {
			return err
		}

		if err = repos.Authorizations.Save(ctx, authz); err != nil {
			return err
		}

		if err = repos.Profiles.Save(ctx, prof); err != nil {
			return err
		}

		return nil
	})
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func (uc *RegisterUseCase) generateUniqueTag(ctx context.Context, repos *Repositories) (profile.Tag, error) {
	const maxAttempts = 5
	for i := 0; i < maxAttempts; i++ {
		suffix, err := randomHex(4)
		if err != nil {
			return "", err
		}
		candidate := "player_" + suffix
		tag, err := profile.NewTag(candidate)
		if err != nil {
			continue
		}
		exists, err := repos.Profiles.ExistsByTag(ctx, tag)
		if err != nil {
			return "", err
		}
		if !exists {
			return tag, nil
		}
	}
	return "", ErrTagGenerationFailed
}
