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

type GameRepository struct {
	client *redis.Client
}

func NewGameRepository(client *redis.Client) *GameRepository {
	return &GameRepository{
		client: client,
	}
}

func gameKey(id uuid.UUID) string {
	return fmt.Sprintf("game:%s", id.String())
}

func (r *GameRepository) Save(ctx context.Context, game *domain.Game) error {
	dto := toDomainDTO(game)
	data, err := json.Marshal(dto)
	if err != nil {
		return err
	}
	key := gameKey(game.ID())
	err = r.client.Set(ctx, key, data, 24*time.Hour).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *GameRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Game, error) {
	key := gameKey(id)

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

func (r *GameRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.client.Del(ctx, gameKey(id)).Err()
}
