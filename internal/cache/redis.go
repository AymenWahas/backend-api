package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type RedisClient struct {
	client *redis.Client
}

func NewRedis(addr string) *RedisClient {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisClient{
		client: client,
	}
}

func (r *RedisClient) Ping(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}
func (r *RedisClient) Get(
	ctx context.Context,
	key string,
	value interface{},
) error {
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}

		return fmt.Errorf("failed to get redis value: %w", err)
	}

	if err := json.Unmarshal(data, value); err != nil {
		return fmt.Errorf("failed to unmarshal redis value: %w", err)
	}

	return nil
}
func (r *RedisClient) Client() *redis.Client {
	return r.client
}
func (r *RedisClient) Set(
	ctx context.Context,
	key string,
	value interface{},
	ttl time.Duration,
) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal redis value: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("failed to set redis value: %w", err)
	}

	return nil
}

func (r *RedisClient) Delete(
	ctx context.Context,
	key string,
) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete redis value: %w", err)
	}

	return nil
}
