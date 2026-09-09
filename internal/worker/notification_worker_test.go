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

func TestNotificationWorkerIntegration(
	t *testing.T,
) {
	cfg := config.Config{
		DBHost:     "localhost",
		DBPort:     "5434",
		DBUser:     "postgres",
		DBPassword: "postgrespassword",
		DBName:     "employee_db",
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		t.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf(
			"failed to get sql.DB: %v",
			err,
		)
	}

	defer sqlDB.Close()

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf(
			"failed to connect to Redis: %v",
			err,
		)
	}

	defer redisClient.Close()

	notificationRepo :=
		postgres.NewNotificationRepository(db)

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			nil,
		),
	)

	worker := NewNotificationWorker(
		redisClient,
		notificationRepo,
		logger,
	)

	// Use unique Redis names so this test does not
	// conflict with the running application worker.
	worker.stream = fmt.Sprintf(
		"test-task-events-%d",
		time.Now().UnixNano(),
	)

	worker.group = fmt.Sprintf(
		"test-notifications-%d",
		time.Now().UnixNano(),
	)

	worker.consumer = "test-worker"

	// Make reclaim fast for tests.
	worker.claimIdle = 100 * time.Millisecond

	defer redisClient.Del(
		context.Background(),
		worker.stream,
		worker.stream+":dlq",
	).Err()

	// --------------------------------------------------
	// Create Consumer Group
	// --------------------------------------------------

	if err := worker.EnsureConsumerGroup(
		ctx,
	); err != nil {
		t.Fatalf(
			"failed to create consumer group: %v",
			err,
		)
	}

	t.Log("consumer group created")

	// --------------------------------------------------
	// Create Task Created Event
	// --------------------------------------------------

	taskEvent := event.TaskCreatedEvent{
		EventID:   fmt.Sprintf("test-event-%d", time.Now().UnixNano()),
		EventType: "task.created",
		TaskID:    999999,
		CreatedAt: time.Now().UTC(),
	}

	eventData, err := json.Marshal(taskEvent)
	if err != nil {
		t.Fatalf(
			"failed to marshal event: %v",
			err,
		)
	}

	// --------------------------------------------------
	// Publish Event
	// --------------------------------------------------

	messageID, err := redisClient.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: worker.stream,
			ID:     "*",
			Values: map[string]interface{}{
				"event": string(eventData),
			},
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to publish event: %v",
			err,
		)
	}

	t.Logf(
		"event published: message_id=%s",
		messageID,
	)

	// --------------------------------------------------
	// Consume Event
	// --------------------------------------------------

	streams, err := redisClient.XReadGroup(
		ctx,
		&redis.XReadGroupArgs{
			Group:    worker.group,
			Consumer: worker.consumer,
			Streams: []string{
				worker.stream,
				">",
			},
			Count: 1,
			Block: 2 * time.Second,
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to consume event: %v",
			err,
		)
	}

	if len(streams) == 0 ||
		len(streams[0].Messages) == 0 {
		t.Fatal("expected one message from Redis Stream")
	}

	message := streams[0].Messages[0]

	t.Logf(
		"event consumed: message_id=%s",
		message.ID,
	)

	// --------------------------------------------------
	// Process Event
	// --------------------------------------------------

	if err := worker.handleMessage(
		ctx,
		message,
	); err != nil {
		t.Fatalf(
			"failed to process event: %v",
			err,
		)
	}

	t.Log("event processed successfully")

	// --------------------------------------------------
	// Verify PostgreSQL
	// --------------------------------------------------

	var count int64

	result := db.Model(
		&domain.Notification{},
	).
		Where(
			"event_id = ?",
			taskEvent.EventID,
		).
		Count(&count)

	if result.Error != nil {
		t.Fatalf(
			"failed to query notification: %v",
			result.Error,
		)
	}

	if count != 1 {
		t.Fatalf(
			"expected exactly 1 notification, got %d",
			count,
		)
	}

	t.Log(
		"PostgreSQL notification created successfully",
	)

	// --------------------------------------------------
	// Idempotency Test
	// --------------------------------------------------

	t.Log(
		"testing duplicate event",
	)

	duplicateMessageID, err := redisClient.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: worker.stream,
			ID:     "*",
			Values: map[string]interface{}{
				"event": string(eventData),
			},
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to publish duplicate event: %v",
			err,
		)
	}

	duplicateStreams, err :=
		redisClient.XReadGroup(
			ctx,
			&redis.XReadGroupArgs{
				Group:    worker.group,
				Consumer: worker.consumer,
				Streams: []string{
					worker.stream,
					">",
				},
				Count: 1,
				Block: 2 * time.Second,
			},
		).Result()

	if err != nil {
		t.Fatalf(
			"failed to consume duplicate event: %v",
			err,
		)
	}

	if len(duplicateStreams) == 0 ||
		len(duplicateStreams[0].Messages) == 0 {
		t.Fatal(
			"expected duplicate message from Redis Stream",
		)
	}

	duplicateMessage :=
		duplicateStreams[0].Messages[0]

	if duplicateMessage.ID != duplicateMessageID {
		t.Fatalf(
			"expected duplicate message ID %s, got %s",
			duplicateMessageID,
			duplicateMessage.ID,
		)
	}

	if err := worker.handleMessage(
		ctx,
		duplicateMessage,
	); err != nil {
		t.Fatalf(
			"duplicate event should be handled safely: %v",
			err,
		)
	}

	// --------------------------------------------------
	// Verify Idempotency
	// --------------------------------------------------

	result = db.Model(
		&domain.Notification{},
	).
		Where(
			"event_id = ?",
			taskEvent.EventID,
		).
		Count(&count)

	if result.Error != nil {
		t.Fatalf(
			"failed to verify duplicate notification: %v",
			result.Error,
		)
	}

	if count != 1 {
		t.Fatalf(
			"idempotency failed: expected 1 notification, got %d",
			count,
		)
	}

	t.Log(
		"idempotency verified: duplicate event did not create another notification",
	)
}

func TestNotificationWorkerRetryAndDLQ(
	t *testing.T,
) {
	cfg := config.Config{
		DBHost:     "localhost",
		DBPort:     "5434",
		DBUser:     "postgres",
		DBPassword: "postgrespassword",
		DBName:     "employee_db",
	}

	db, err := database.NewPostgres(cfg)
	if err != nil {
		t.Fatalf(
			"failed to connect to PostgreSQL: %v",
			err,
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf(
			"failed to get sql.DB: %v",
			err,
		)
	}

	defer sqlDB.Close()

	redisClient := redis.NewClient(
		&redis.Options{
			Addr: "localhost:6379",
		},
	)

	ctx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		t.Fatalf(
			"failed to connect to Redis: %v",
			err,
		)
	}

	defer redisClient.Close()

	notificationRepo :=
		postgres.NewNotificationRepository(db)

	logger := slog.New(
		slog.NewTextHandler(
			os.Stdout,
			nil,
		),
	)

	worker := NewNotificationWorker(
		redisClient,
		notificationRepo,
		logger,
	)

	// Use unique Redis names for this test.
	worker.stream = fmt.Sprintf(
		"test-retry-task-events-%d",
		time.Now().UnixNano(),
	)

	worker.group = fmt.Sprintf(
		"test-retry-notifications-%d",
		time.Now().UnixNano(),
	)

	worker.consumer = "test-retry-worker"

	// Make retries fast for the test.
	worker.claimIdle = 100 * time.Millisecond

	defer redisClient.Del(
		context.Background(),
		worker.stream,
		worker.stream+":dlq",
	).Err()

	if err := worker.EnsureConsumerGroup(
		ctx,
	); err != nil {
		t.Fatalf(
			"failed to create consumer group: %v",
			err,
		)
	}

	t.Log("retry test consumer group created")

	// --------------------------------------------------
	// Publish a deliberately invalid event.
	//
	// JSON is invalid, so processMessage will fail.
	// --------------------------------------------------

	messageID, err := redisClient.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: worker.stream,
			ID:     "*",
			Values: map[string]interface{}{
				"event": "{invalid-json",
			},
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to publish invalid event: %v",
			err,
		)
	}

	t.Logf(
		"invalid event published: message_id=%s",
		messageID,
	)

	// --------------------------------------------------
	// First delivery
	// --------------------------------------------------

	streams, err := redisClient.XReadGroup(
		ctx,
		&redis.XReadGroupArgs{
			Group:    worker.group,
			Consumer: worker.consumer,
			Streams: []string{
				worker.stream,
				">",
			},
			Count: 1,
			Block: 2 * time.Second,
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to consume invalid event: %v",
			err,
		)
	}

	if len(streams) == 0 ||
		len(streams[0].Messages) == 0 {
		t.Fatal(
			"expected invalid event from Redis Stream",
		)
	}

	message := streams[0].Messages[0]

	if message.ID != messageID {
		t.Fatalf(
			"expected message ID %s, got %s",
			messageID,
			message.ID,
		)
	}

	// First processing attempt must fail.
	err = worker.handleMessage(
		ctx,
		message,
	)

	if err == nil {
		t.Fatal(
			"expected first processing attempt to fail",
		)
	}

	t.Logf(
		"attempt 1 failed as expected: %v",
		err,
	)

	// --------------------------------------------------
	// Verify message is still Pending.
	// --------------------------------------------------

	pending, err := redisClient.XPendingExt(
		ctx,
		&redis.XPendingExtArgs{
			Stream: worker.stream,
			Group:  worker.group,
			Start:  "-",
			End:    "+",
			Count:  10,
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to inspect pending messages: %v",
			err,
		)
	}

	if len(pending) != 1 {
		t.Fatalf(
			"expected 1 pending message, got %d",
			len(pending),
		)
	}

	if pending[0].ID != messageID {
		t.Fatalf(
			"expected pending message %s, got %s",
			messageID,
			pending[0].ID,
		)
	}

	t.Log(
		"attempt 1 verified: message remains pending",
	)

	// --------------------------------------------------
	// Attempt 2
	// --------------------------------------------------

	time.Sleep(
		worker.claimIdle + 50*time.Millisecond,
	)

	if err := worker.reclaimPendingMessages(
		ctx,
	); err != nil {
		t.Fatalf(
			"failed to reclaim message for attempt 2: %v",
			err,
		)
	}

	attempts, err := redisClient.Get(
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

	t.Log(
		"attempt 2 verified",
	)

	// --------------------------------------------------
	// Attempt 3
	// --------------------------------------------------

	time.Sleep(
		worker.claimIdle + 50*time.Millisecond,
	)

	if err := worker.reclaimPendingMessages(
		ctx,
	); err != nil {
		t.Fatalf(
			"failed to reclaim message for attempt 3: %v",
			err,
		)
	}

	// --------------------------------------------------
	// Verify DLQ
	// --------------------------------------------------

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
			"expected exactly 1 DLQ message, got %d",
			len(dlqMessages),
		)
	}

	dlqMessage := dlqMessages[0]

	originalID, ok :=
		dlqMessage.Values["original_message_id"].(string)

	if !ok {
		t.Fatal(
			"DLQ message is missing original_message_id",
		)
	}

	if originalID != messageID {
		t.Fatalf(
			"expected original message ID %s in DLQ, got %s",
			messageID,
			originalID,
		)
	}

	dlqAttempts, ok :=
		dlqMessage.Values["attempts"].(string)

	if !ok {
		// Redis may return numeric values as strings,
		// but keep the test tolerant of integer values.
		switch value := dlqMessage.Values["attempts"].(type) {
		case int64:
			dlqAttempts = fmt.Sprintf(
				"%d",
				value,
			)

		case int:
			dlqAttempts = fmt.Sprintf(
				"%d",
				value,
			)

		default:
			t.Fatalf(
				"unexpected DLQ attempts type: %T",
				value,
			)
		}
	}

	if dlqAttempts != "3" {
		t.Fatalf(
			"expected DLQ attempts=3, got %s",
			dlqAttempts,
		)
	}

	t.Log(
		"DLQ verified: message moved after 3 attempts",
	)

	// --------------------------------------------------
	// Verify original message was ACKed.
	// --------------------------------------------------

	pending, err = redisClient.XPendingExt(
		ctx,
		&redis.XPendingExtArgs{
			Stream: worker.stream,
			Group:  worker.group,
			Start:  "-",
			End:    "+",
			Count:  10,
		},
	).Result()

	if err != nil {
		t.Fatalf(
			"failed to inspect pending messages after DLQ: %v",
			err,
		)
	}

	if len(pending) != 0 {
		t.Fatalf(
			"expected original message to be ACKed after DLQ, got %d pending messages",
			len(pending),
		)
	}

	t.Log(
		"ACK verified: original message removed from Pending Entries",
	)

	// --------------------------------------------------
	// Verify retry counter was cleared.
	// --------------------------------------------------

	_, err = redisClient.Get(
		ctx,
		worker.retryKey(messageID),
	).Result()

	if !errors.Is(err, redis.Nil) {
		t.Fatalf(
			"expected retry counter to be deleted, got error: %v",
			err,
		)
	}

	t.Log(
		"retry counter cleanup verified",
	)
}
