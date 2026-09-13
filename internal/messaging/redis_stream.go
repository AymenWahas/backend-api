package messaging

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"

	"backend-api/internal/event"
)

const TaskEventsStream = "task-events"

type RedisStreamPublisher struct {
	client *redis.Client
}

func NewRedisStreamPublisher(
	client *redis.Client,
) *RedisStreamPublisher {
	return &RedisStreamPublisher{
		client: client,
	}
}

func (p *RedisStreamPublisher) PublishTaskCreated(
	ctx context.Context,
	e event.TaskCreatedEvent,
) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}

	_, err = p.client.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: TaskEventsStream,
			ID:     "*", // redias will generate a unique ID for the message
			Values: map[string]interface{}{
				"event": string(data),
			},
		},
	).Result()

	return err
}
