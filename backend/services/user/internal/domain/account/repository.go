package account

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Save(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id uuid.UUID) (*Account, error)
}
