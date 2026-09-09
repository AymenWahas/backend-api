package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"backend-api/internal/domain"
	"backend-api/internal/event"
	"backend-api/internal/repository"
)

const (
	TaskEventsStream    = "task-events"
	TaskEventsDLQStream = "task-events:dlq"

	ConsumerGroup = "notifications"
	ConsumerName  = "notification-worker-1"

	MaxAttempts = 3
	ClaimIdle   = 10 * time.Second

	RetryKeyPrefix = "task-events:retry:"
)

type NotificationWorker struct {
	redis            *redis.Client
	notificationRepo repository.NotificationRepository
	logger           *slog.Logger

	stream      string
	group       string
	consumer    string
	maxAttempts int
	claimIdle   time.Duration
}

func NewNotificationWorker(
	redisClient *redis.Client,
	notificationRepo repository.NotificationRepository,
	logger *slog.Logger,
) *NotificationWorker {
	return &NotificationWorker{
		redis:            redisClient,
		notificationRepo: notificationRepo,
		logger:           logger,

		stream:      TaskEventsStream,
		group:       ConsumerGroup,
		consumer:    ConsumerName,
		maxAttempts: MaxAttempts,
		claimIdle:   ClaimIdle,
	}
}

func (w *NotificationWorker) EnsureConsumerGroup(
	ctx context.Context,
) error {
	err := w.redis.XGroupCreateMkStream(
		ctx,
		w.stream,
		w.group,
		"$",
	).Err()

	if err != nil &&
		err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf(
			"failed to create consumer group: %w",
			err,
		)
	}

	return nil
}

func (w *NotificationWorker) retryKey(
	messageID string,
) string {
	return RetryKeyPrefix + messageID
}

func (w *NotificationWorker) incrementAttempts(
	ctx context.Context,
	messageID string,
) (int64, error) {
	attempts, err := w.redis.Incr(
		ctx,
		w.retryKey(messageID),
	).Result()

	if err != nil {
		return 0, fmt.Errorf(
			"failed to increment retry attempts: %w",
			err,
		)
	}

	return attempts, nil
}

func (w *NotificationWorker) clearAttempts(
	ctx context.Context,
	messageID string,
) error {
	return w.redis.Del(
		ctx,
		w.retryKey(messageID),
	).Err()
}

func (w *NotificationWorker) processMessage(
	ctx context.Context,
	message redis.XMessage,
) error {
	rawEvent, ok := message.Values["event"].(string)

	if !ok {
		return fmt.Errorf(
			"event field missing in message %s",
			message.ID,
		)
	}

	var taskEvent event.TaskCreatedEvent

	if err := json.Unmarshal(
		[]byte(rawEvent),
		&taskEvent,
	); err != nil {
		return fmt.Errorf(
			"failed to decode task.created event: %w",
			err,
		)
	}

	notification := &domain.Notification{
		EventID: taskEvent.EventID,
		TaskID:  taskEvent.TaskID,
		Message: fmt.Sprintf(
			"Task %d was created",
			taskEvent.TaskID,
		),
	}

	if err := w.notificationRepo.Create(
		ctx,
		notification,
	); err != nil {
		if errors.Is(
			err,
			domain.ErrNotificationAlreadyExists,
		) {
			w.logger.Info(
				"duplicate notification skipped",
				"event_id",
				taskEvent.EventID,
				"task_id",
				taskEvent.TaskID,
				"message_id",
				message.ID,
			)

			return nil
		}

		return err
	}

	return nil
}

func (w *NotificationWorker) sendToDLQ(
	ctx context.Context,
	message redis.XMessage,
	attempts int64,
) error {
	rawEvent, ok := message.Values["event"].(string)

	if !ok {
		return fmt.Errorf(
			"event field missing for DLQ message %s",
			message.ID,
		)
	}

	_, err := w.redis.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: w.stream + ":dlq",
			ID:     "*",
			Values: map[string]interface{}{
				"event":               rawEvent,
				"original_message_id": message.ID,
				"attempts":            attempts,
			},
		},
	).Result()

	if err != nil {
		return fmt.Errorf(
			"failed to publish message to DLQ: %w",
			err,
		)
	}

	return nil
}

func (w *NotificationWorker) handleMessage(
	ctx context.Context,
	message redis.XMessage,
) error {
	err := w.processMessage(
		ctx,
		message,
	)

	if err == nil {
		if err := w.redis.XAck(
			ctx,
			w.stream,
			w.group,
			message.ID,
		).Err(); err != nil {
			return fmt.Errorf(
				"failed to ACK message %s: %w",
				message.ID,
				err,
			)
		}

		if err := w.clearAttempts(
			ctx,
			message.ID,
		); err != nil {
			w.logger.Warn(
				"failed to clear retry counter",
				"message_id",
				message.ID,
				"error",
				err,
			)
		}

		w.logger.Info(
			"notification event processed",
			"message_id",
			message.ID,
		)

		return nil
	}

	attempts, attemptsErr := w.incrementAttempts(
		ctx,
		message.ID,
	)

	if attemptsErr != nil {
		return fmt.Errorf(
			"processing failed: %w; retry tracking failed: %v",
			err,
			attemptsErr,
		)
	}

	w.logger.Error(
		"failed to process message",
		"message_id",
		message.ID,
		"attempt",
		attempts,
		"max_attempts",
		w.maxAttempts,
		"error",
		err,
	)

	if attempts < int64(w.maxAttempts) {
		// Do NOT ACK.
		//
		// The message remains pending in Redis.
		// XAUTOCLAIM will reclaim it later.
		return err
	}

	// Maximum attempts reached.
	// Move the message to the DLQ first.
	if err := w.sendToDLQ(
		ctx,
		message,
		attempts,
	); err != nil {
		// Do NOT ACK if DLQ publishing failed.
		return err
	}

	// DLQ succeeded, so ACK the original message.
	if err := w.redis.XAck(
		ctx,
		w.stream,
		w.group,
		message.ID,
	).Err(); err != nil {
		return fmt.Errorf(
			"failed to ACK DLQ message %s: %w",
			message.ID,
			err,
		)
	}

	if err := w.clearAttempts(
		ctx,
		message.ID,
	); err != nil {
		w.logger.Warn(
			"failed to clear DLQ retry counter",
			"message_id",
			message.ID,
			"error",
			err,
		)
	}

	w.logger.Error(
		"message moved to DLQ",
		"message_id",
		message.ID,
		"attempts",
		attempts,
		"dlq_stream",
		TaskEventsDLQStream,
	)

	return nil
}

func (w *NotificationWorker) processNewMessages(
	ctx context.Context,
) error {
	streams, err := w.redis.XReadGroup(
		ctx,
		&redis.XReadGroupArgs{
			Group:    w.group,
			Consumer: w.consumer,
			Streams:  []string{w.stream, ">"},
			Count:    10,
			Block:    2 * time.Second,
		},
	).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil
		}

		if errors.Is(err, context.Canceled) {
			return ctx.Err()
		}

		return fmt.Errorf(
			"failed to read task events: %w",
			err,
		)
	}

	for _, stream := range streams {
		for _, message := range stream.Messages {
			if err := w.handleMessage(
				ctx,
				message,
			); err != nil {
				w.logger.Error(
					"message handling failed",
					"message_id",
					message.ID,
					"error",
					err,
				)
			}
		}
	}

	return nil
}
func (w *NotificationWorker) reclaimPendingMessages(
	ctx context.Context,
) error {
	startID := "0-0"

	for {
		messages, nextID, err := w.redis.XAutoClaim(
			ctx,
			&redis.XAutoClaimArgs{
				Stream:   w.stream,
				Group:    w.group,
				Consumer: w.consumer,
				MinIdle:  w.claimIdle,
				Start:    startID,
				Count:    10,
			},
		).Result()

		if err != nil {
			if errors.Is(err, redis.Nil) {
				return nil
			}

			if errors.Is(err, context.Canceled) {
				return ctx.Err()
			}

			return fmt.Errorf(
				"failed to auto-claim pending messages: %w",
				err,
			)
		}

		if len(messages) == 0 {
			return nil
		}

		for _, message := range messages {
			if err := w.handleMessage(
				ctx,
				message,
			); err != nil {
				w.logger.Error(
					"reclaimed message handling failed",
					"message_id",
					message.ID,
					"error",
					err,
				)
			}
		}

		startID = nextID

		if startID == "0-0" {
			return nil
		}
	}
}
func (w *NotificationWorker) Run(
	ctx context.Context,
) error {
	if err := w.EnsureConsumerGroup(ctx); err != nil {
		return err
	}

	w.logger.Info(
		"notification worker started",
		"stream",
		w.stream,
		"group",
		w.group,
		"consumer",
		w.consumer,
	)

	ticker := time.NewTicker(
		w.claimIdle / 2,
	)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info(
				"notification worker stopped",
			)

			return ctx.Err()

		case <-ticker.C:
			if err := w.reclaimPendingMessages(ctx); err != nil {
				w.logger.Error(
					"failed to reclaim pending messages",
					"error",
					err,
				)
			}

		default:
			if err := w.processNewMessages(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return ctx.Err()
				}

				w.logger.Error(
					"failed to process new messages",
					"error",
					err,
				)
			}
		}
	}
}
