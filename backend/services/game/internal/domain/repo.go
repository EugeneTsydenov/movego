package domain

import (
	"context"
)

type GameRepo interface {
	Save(ctx context.Context, game *Game) error
	GetByID(ctx context.Context, id GameID) (*Game, error)
	Delete(ctx context.Context, id GameID) error
}
