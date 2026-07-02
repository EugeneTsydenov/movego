package authorization

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, auth *Authorization) error
	FindByAccountID(ctx context.Context, accountID uuid.UUID) (*Authorization, error)
	CountByRole(ctx context.Context, role Role) (int, error)
}
