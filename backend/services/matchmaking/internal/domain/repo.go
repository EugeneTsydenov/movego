package domain

import "context"

type QueueRepo interface {
	Save(ctx context.Context, player *Player) error
	FindAll(ctx context.Context) ([]*Player, error)
	Delete(ctx context.Context, ids ...PlayerID) error
}
