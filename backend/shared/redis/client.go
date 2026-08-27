package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, cfg *redis.Options) (*redis.Client, error) {
	client := redis.NewClient(cfg)

	if err := redisotel.InstrumentTracing(client); err != nil {
		return nil, fmt.Errorf("failed to instrument redis tracing: %w", err)
	}

	if err := redisotel.InstrumentMetrics(client); err != nil {
		return nil, fmt.Errorf("failed to instrument redis metrics: %w", err)
	}

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return client, nil
}
