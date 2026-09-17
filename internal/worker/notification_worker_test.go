package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"backend-api/internal/config"
	"backend-api/internal/database"
	"backend-api/internal/domain"
	"backend-api/internal/event"
	"backend-api/internal/repository/postgres"
)

type timeoutNotificationRepository struct{}

func (r *timeoutNotificationRepository) Create(
	ctx context.Context,
	notification *domain.Notification,
) error {
	<-ctx.Done()
	return ctx.Err()
}

type failingNotificationRepository struct{}

func (r *failingNotificationRepository) Create(
	ctx context.Context,
	notification *domain.Notification,
) error {
	return errors.New("simulated notification repository failure")
}

func TestNotificationWorkerIntegration(t *testing.T) {
	t.Setenv(
		"JWT_SECRET",
		"test-secret",
	)
	ctx := context.Background()

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)
	defer redisClient.Close()

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			nil,
		),
	)

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf(
			"failed to load config: %v",
			err,
		)
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		t.Fatalf(
			"failed to connect to postgres: %v",
			err,
		)
	}

	notificationRepo := postgres.NewNotificationRepository(db)

	worker := NewNotificationWorker(
		redisClient,
		notificationRepo,
		logger,
	)

	worker.stream = fmt.Sprintf(
		"test-task-events-%d",
		time.Now().UnixNano(),
	)

	worker.group = fmt.Sprintf(
		"test-notifications-%d",
		time.Now().UnixNano(),
	)

	worker.consumer = "test-worker"

	defer redisClient.Del(
		context.Background(),
		worker.stream,
		worker.stream+":dlq",
	)

	if err := worker.EnsureConsumerGroup(ctx); err != nil {
		t.Fatalf(
			"failed to create consumer group: %v",
			err,
		)
	}

	eventID := fmt.Sprintf(
		"test-event-%d",
		time.Now().UnixNano(),
	)

	taskID := uint(1)

	testEvent := event.TaskCreatedEvent{
		EventID:   eventID,
		EventType: "task.created",
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}

	payload, err := json.Marshal(testEvent)
	if err != nil {
		t.Fatalf(
			"failed to marshal event: %v",
			err,
		)
	}

	_, err = redisClient.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: worker.stream,
			Values: map[string]interface{}{
				"event": string(payload),
			},
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to add event to stream: %v",
			err,
		)
	}

	if err := worker.processNewMessages(ctx); err != nil {
		t.Fatalf(
			"failed to process message: %v",
			err,
		)
	}
}

func TestNotificationWorkerRetryAndDLQ(t *testing.T) {
	ctx := context.Background()

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)
	defer redisClient.Close()

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			nil,
		),
	)

	notificationRepo := &failingNotificationRepository{}

	worker := NewNotificationWorker(
		redisClient,
		notificationRepo,
		logger,
	)

	worker.stream = fmt.Sprintf(
		"test-retry-task-events-%d",
		time.Now().UnixNano(),
	)

	worker.group = fmt.Sprintf(
		"test-retry-notifications-%d",
		time.Now().UnixNano(),
	)

	worker.consumer = "test-retry-worker"

	// Production value is 10 seconds.
	// For the test we make pending messages
	// eligible for XAUTOCLAIM quickly.
	worker.claimIdle = 50 * time.Millisecond

	defer redisClient.Del(
		context.Background(),
		worker.stream,
		worker.stream+":dlq",
	)

	if err := worker.EnsureConsumerGroup(ctx); err != nil {
		t.Fatalf(
			"failed to create consumer group: %v",
			err,
		)
	}

	eventID := fmt.Sprintf(
		"retry-event-%d",
		time.Now().UnixNano(),
	)

	testEvent := event.TaskCreatedEvent{
		EventID:   eventID,
		EventType: "task.created",
		TaskID:    999,
		CreatedAt: time.Now().UTC(),
	}

	payload, err := json.Marshal(testEvent)
	if err != nil {
		t.Fatalf(
			"failed to marshal event: %v",
			err,
		)
	}

	messageID, err := redisClient.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: worker.stream,
			Values: map[string]interface{}{
				"event": string(payload),
			},
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to add event to stream: %v",
			err,
		)
	}

	// Attempt 1:
	// The repository fails.
	// The retry counter becomes 1.
	// The message remains pending.
	if err := worker.processNewMessages(ctx); err != nil {
		t.Fatalf(
			"unexpected error while processing attempt 1: %v",
			err,
		)
	}

	attempts, err := redisClient.Get(
		ctx,
		worker.retryKey(messageID),
	).Int64()

	if err != nil {
		t.Fatalf(
			"failed to read retry counter after attempt 1: %v",
			err,
		)
	}

	if attempts != 1 {
		t.Fatalf(
			"expected 1 attempt, got %d",
			attempts,
		)
	}

	// Attempt 2:
	// Wait until the message becomes idle.
	time.Sleep(
		worker.claimIdle + 20*time.Millisecond,
	)

	if err := worker.reclaimPendingMessages(ctx); err != nil {
		t.Fatalf(
			"unexpected error while reclaiming attempt 2: %v",
			err,
		)
	}

	attempts, err = redisClient.Get(
		ctx,
		worker.retryKey(messageID),
	).Int64()

	if err != nil {
		t.Fatalf(
			"failed to read retry counter after attempt 2: %v",
			err,
		)
	}

	if attempts != 2 {
		t.Fatalf(
			"expected 2 attempts, got %d",
			attempts,
		)
	}

	// Attempt 3:
	// The maximum is 3.
	// The third failure must move the message to the DLQ.
	time.Sleep(
		worker.claimIdle + 20*time.Millisecond,
	)

	if err := worker.reclaimPendingMessages(ctx); err != nil {
		t.Fatalf(
			"unexpected error while reclaiming attempt 3: %v",
			err,
		)
	}

	// Verify retry counter was cleared.
	retryExists, err := redisClient.Exists(
		ctx,
		worker.retryKey(messageID),
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to check retry key: %v",
			err,
		)
	}

	if retryExists != 0 {
		t.Fatalf(
			"expected retry counter to be cleared, key still exists",
		)
	}

	// Verify the message exists in the DLQ.
	dlqMessages, err := redisClient.XRange(
		ctx,
		worker.stream+":dlq",
		"-",
		"+",
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to read DLQ: %v",
			err,
		)
	}

	if len(dlqMessages) != 1 {
		t.Fatalf(
			"expected 1 message in DLQ, got %d",
			len(dlqMessages),
		)
	}

	dlqMessage := dlqMessages[0]

	if dlqMessage.Values["original_message_id"] != messageID {
		t.Fatalf(
			"expected original message ID %s in DLQ, got %v",
			messageID,
			dlqMessage.Values["original_message_id"],
		)
	}

	attemptValue := fmt.Sprint(
		dlqMessage.Values["attempts"],
	)

	if attemptValue != "3" {
		t.Fatalf(
			"expected DLQ attempts to be 3, got %s",
			attemptValue,
		)
	}

	// Verify the original message was ACKed.
	pending, err := redisClient.XPending(
		ctx,
		worker.stream,
		worker.group,
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to inspect pending messages: %v",
			err,
		)
	}

	if pending.Count != 0 {
		t.Fatalf(
			"expected 0 pending messages after DLQ, got %d",
			pending.Count,
		)
	}

	t.Log(
		"retry and DLQ behavior verified successfully",
	)
}

func TestNotificationWorkerCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)
	defer redisClient.Close()

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			nil,
		),
	)

	notificationRepo := &timeoutNotificationRepository{}

	worker := NewNotificationWorker(
		redisClient,
		notificationRepo,
		logger,
	)

	worker.stream = fmt.Sprintf(
		"test-cancel-task-events-%d",
		time.Now().UnixNano(),
	)

	worker.group = fmt.Sprintf(
		"test-cancel-notifications-%d",
		time.Now().UnixNano(),
	)

	worker.consumer = "test-cancel-worker"

	worker.claimIdle = 100 * time.Millisecond

	defer redisClient.Del(
		context.Background(),
		worker.stream,
		worker.stream+":dlq",
	)

	if err := worker.EnsureConsumerGroup(ctx); err != nil {
		t.Fatalf(
			"failed to create consumer group: %v",
			err,
		)
	}

	done := make(chan error, 1)

	go func() {
		done <- worker.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	cancel()

	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf(
				"expected context.Canceled, got %v",
				err,
			)
		}

		t.Log(
			"worker cancellation verified successfully",
		)

	case <-time.After(2 * time.Second):
		t.Fatal(
			"worker did not stop after context cancellation",
		)
	}
}
