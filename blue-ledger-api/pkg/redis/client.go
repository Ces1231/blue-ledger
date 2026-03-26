package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// NewClient creates and validates a Redis client from the given URL.
func NewClient(redisURL string) (*redis.Client, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

// Ping verifies that the Redis connection is alive. Used by /healthz.
func Ping(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
