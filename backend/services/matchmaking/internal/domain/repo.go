package domain

import "context"

type QueueRepo interface {
	Save(ctx context.Context, player *Player) error
	Exists(ctx context.Context, playerID PlayerID) (bool, error)
	FindAll(ctx context.Context) ([]*Player, error)
	Delete(ctx context.Context, ids ...PlayerID) error
}
