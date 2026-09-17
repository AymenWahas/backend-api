package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"backend-api/internal/observability"
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
	start := time.Now()

	err := r.client.Ping(ctx).Err()

	observability.RedisOperations.
		WithLabelValues("ping").
		Inc()

	observability.RedisOperationDuration.
		WithLabelValues("ping").
		Observe(time.Since(start).Seconds())

	if err != nil {
		observability.RedisErrors.
			WithLabelValues("ping").
			Inc()

		return fmt.Errorf("redis ping failed: %w", err)
	}

	return nil
}

func (r *RedisClient) Get(
	ctx context.Context,
	key string,
	value interface{},
) error {
	start := time.Now()

	data, err := r.client.Get(ctx, key).Bytes()

	observability.RedisOperations.
		WithLabelValues("get").
		Inc()

	observability.RedisOperationDuration.
		WithLabelValues("get").
		Observe(time.Since(start).Seconds())

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}

		observability.RedisErrors.
			WithLabelValues("get").
			Inc()

		return fmt.Errorf("failed to get redis value: %w", err)
	}

	if err := json.Unmarshal(data, value); err != nil {
		observability.RedisErrors.
			WithLabelValues("get").
			Inc()

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
	start := time.Now()

	data, err := json.Marshal(value)

	if err != nil {
		observability.RedisErrors.
			WithLabelValues("set").
			Inc()

		return fmt.Errorf("failed to marshal redis value: %w", err)
	}

	err = r.client.Set(ctx, key, data, ttl).Err()

	observability.RedisOperations.
		WithLabelValues("set").
		Inc()

	observability.RedisOperationDuration.
		WithLabelValues("set").
		Observe(time.Since(start).Seconds())

	if err != nil {
		observability.RedisErrors.
			WithLabelValues("set").
			Inc()

		return fmt.Errorf("failed to set redis value: %w", err)
	}

	return nil
}

func (r *RedisClient) Delete(
	ctx context.Context,
	key string,
) error {
	start := time.Now()

	err := r.client.Del(ctx, key).Err()

	observability.RedisOperations.
		WithLabelValues("delete").
		Inc()

	observability.RedisOperationDuration.
		WithLabelValues("delete").
		Observe(time.Since(start).Seconds())

	if err != nil {
		observability.RedisErrors.
			WithLabelValues("delete").
			Inc()

		return fmt.Errorf("failed to delete redis value: %w", err)
	}

	return nil
}
