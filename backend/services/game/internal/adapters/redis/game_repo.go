package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"game/internal/domain"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type GameRepo struct {
	client *redis.Client
}

func NewGameRepo(client *redis.Client) *GameRepo {
	return &GameRepo{
		client: client,
	}
}

func gameKey(id uuid.UUID) string {
	return fmt.Sprintf("game:%s", id.String())
}

func (r *GameRepo) Save(ctx context.Context, game *domain.Game) error {
	dto := toGameDTO(game)
	data, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	key := gameKey(game.ID().UUID())
	err = r.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *GameRepo) GetByID(ctx context.Context, id domain.GameID) (*domain.Game, error) {
	key := gameKey(id.UUID())

	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, domain.ErrGameNotFound
		}
		return nil, err
	}

	var dto gameDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return nil, err
	}

	game, err := toDomainGame(dto)
	if err != nil {
		return nil, err
	}

	return game, nil
}

func (r *GameRepo) Delete(ctx context.Context, id domain.GameID) error {
	return r.client.Del(ctx, gameKey(id.UUID())).Err()
}
