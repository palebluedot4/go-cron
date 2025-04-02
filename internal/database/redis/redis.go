package redis

import (
	"context"
	"fmt"

	"go-cron/internal/config"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func New(ctx context.Context, cfg *config.Config) (*Redis, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Storage.Redis.Host, cfg.Storage.Redis.Port),
		Password:     cfg.Storage.Redis.Password,
		DB:           cfg.Storage.Redis.DB,
		DialTimeout:  cfg.Storage.Redis.DialTimeout,
		ReadTimeout:  cfg.Storage.Redis.ReadTimeout,
		WriteTimeout: cfg.Storage.Redis.WriteTimeout,
		MaxRetries:   cfg.Storage.Redis.MaxRetries,
		PoolSize:     cfg.Storage.Redis.PoolSize,
		MinIdleConns: cfg.Storage.Redis.MinIdleConns,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{
		client: client,
	}, nil
}

func (r *Redis) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return r.client.Close()
	}
}

func (r *Redis) IsConnected(ctx context.Context) bool {
	return r.client.Ping(ctx).Err() == nil
}
