package profile

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, p *Profile) error
	FindByAccountID(ctx context.Context, accountID uuid.UUID) (*Profile, error)
	FindByTag(ctx context.Context, tag Tag) (*Profile, error)
	ExistsByTag(ctx context.Context, tag Tag) (bool, error)
}
