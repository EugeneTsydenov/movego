package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/movego/services/user/internal/domain/profile"
)

type UpdateProfileCommand struct {
	AccountID   uuid.UUID
	Tag         *string
	DisplayName *string
	AvatarUrl   *string
	Country     *string
}

type UpdateProfileUseCase struct {
	profiles profile.Repository
}

func NewUpdateProfileUseCase(profiles profile.Repository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		profiles: profiles,
	}
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, cmd UpdateProfileCommand) error {
	profileEntity, err := uc.profiles.FindByAccountID(ctx, cmd.AccountID)
	if err != nil {
		return err
	}

	var updateErrs []error

	if cmd.Tag != nil {
		if err = profileEntity.UpdateTag(*cmd.Tag); err != nil {
			updateErrs = append(updateErrs, fmt.Errorf("update tag: %w", err))
		}
	}

	if cmd.DisplayName != nil {
		if err = profileEntity.UpdateDisplayName(*cmd.DisplayName); err != nil {
			updateErrs = append(updateErrs, fmt.Errorf("update display name: %w", err))
		}
	}

	if cmd.AvatarUrl != nil {
		if err = profileEntity.UpdateAvatarURL(*cmd.AvatarUrl); err != nil {
			updateErrs = append(updateErrs, fmt.Errorf("update avatar url: %w", err))
		}
	}

	if cmd.Country != nil {
		if err = profileEntity.UpdateCountry(*cmd.Country); err != nil {
			updateErrs = append(updateErrs, fmt.Errorf("update country: %w", err))
		}
	}

	if len(updateErrs) > 0 {
		return errors.Join(updateErrs...)
	}

	if err = uc.profiles.Save(ctx, profileEntity); err != nil {
		return fmt.Errorf("save profile: %w", err)
	}

	return nil
}
