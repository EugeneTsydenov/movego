package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"matchmaking/internal/domain"

	"github.com/redis/go-redis/v9"
)

const (
	queueKey = "matchmaking:queue"
)

type QueueRepo struct {
	client *redis.Client
}

func NewQueueRepo(client *redis.Client) *QueueRepo {
	return &QueueRepo{
		client: client,
	}
}

func playerKey(id string) string {
	return fmt.Sprintf("player:%s", id)
}

func (r *QueueRepo) Save(ctx context.Context, player *domain.Player) error {
	dto := toPlayerDTO(player)
	data, err := json.Marshal(dto)
	if err != nil {
		return fmt.Errorf("marshal dto failed: %w", err)
	}

	playerIDStr := player.ID().String()
	pKey := playerKey(playerIDStr)
	score := float64(player.JoinedAt().Unix())

	pipe := r.client.Pipeline()
	pipe.Set(ctx, pKey, data, time.Hour*5)
	pipe.ZAdd(ctx, queueKey, redis.Z{
		Score:  score,
		Member: playerIDStr,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("pipeline exec failed: %w", err)
	}

	return nil
}

func (r *QueueRepo) Exists(ctx context.Context, playerID domain.PlayerID) (bool, error) {
	pKey := playerKey(playerID.String())

	val, err := r.client.Exists(ctx, pKey).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check player existence in redis: %w", err)
	}

	return val > 0, nil
}

func (r *QueueRepo) FindAll(ctx context.Context) ([]*domain.Player, error) {
	ids, err := r.client.ZRange(ctx, queueKey, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = playerKey(id)
	}

	vals, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var players []*domain.Player
	for _, val := range vals {
		if val == nil {
			continue
		}

		strVal, ok := val.(string)
		if !ok {
			continue
		}

		var dto playerDTO
		if err := json.Unmarshal([]byte(strVal), &dto); err != nil {
			return nil, err
		}

		player, err := toDomainPlayer(dto)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	return players, nil
}

func (r *QueueRepo) Delete(ctx context.Context, ids ...domain.PlayerID) error {
	if len(ids) == 0 {
		return nil
	}
	zremArgs := make([]any, len(ids))
	playerKeys := make([]string, len(ids))

	for i, id := range ids {
		idStr := id.String()
		zremArgs[i] = idStr
		playerKeys[i] = playerKey(idStr)
	}

	pipe := r.client.Pipeline()
	pipe.ZRem(ctx, queueKey, zremArgs...)
	pipe.Del(ctx, playerKeys...)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete players in pipeline: %w", err)
	}

	return nil
}
