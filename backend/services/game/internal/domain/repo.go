package domain

import (
	"context"

	"github.com/google/uuid"
)

type GameRepository interface {
	Save(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id uuid.UUID) (*Game, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
