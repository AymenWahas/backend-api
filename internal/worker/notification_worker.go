package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"backend-api/internal/domain"
	"backend-api/internal/event"
	"backend-api/internal/observability"
	"backend-api/internal/repository"
)

const (
	TaskEventsStream    = "task-events"
	TaskEventsDLQStream = "task-events:dlq"

	ConsumerGroup = "notifications"
	ConsumerName  = "notification-worker-1"

	MaxAttempts = 3

	// XAUTOCLAIM will recover messages that were
	// pending for at least this amount of time.
	ClaimIdle = 10 * time.Second

	MessageTimeout = 5 * time.Second
	RetryKeyPrefix = "task-events:retry:" //number of retry

	// Number of concurrent workers processing one batch.
	WorkerCount = 3

	// Maximum number of messages read from Redis at once.
	BatchSize = 10

	// Retry backoff configuration.
	InitialBackoff = 1 * time.Second
	MaxBackoff     = 30 * time.Second
)

type NotificationWorker struct {
	redis            *redis.Client
	notificationRepo repository.NotificationRepository
	logger           *slog.Logger

	stream   string
	group    string
	consumer string

	maxAttempts int
	claimIdle   time.Duration

	workerCount int
	batchSize   int

	initialBackoff time.Duration
	maxBackoff     time.Duration
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

		stream:   TaskEventsStream,
		group:    ConsumerGroup,
		consumer: ConsumerName,

		maxAttempts: MaxAttempts,
		claimIdle:   ClaimIdle,

		workerCount: WorkerCount,
		batchSize:   BatchSize,

		initialBackoff: InitialBackoff,
		maxBackoff:     MaxBackoff,
	}
}

// EnsureConsumerGroup creates the Redis Stream consumer group.
//
// BUSYGROUP means the group already exists,
// which is not an error for our application.
func (w *NotificationWorker) EnsureConsumerGroup(
	ctx context.Context,
) error {
	err := w.redis.XGroupCreateMkStream(
		ctx,
		w.stream,
		w.group,
		"$", // grop of new message start
	).Err()
	//if group mojodh
	if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
		return fmt.Errorf(
			"create consumer group: %w",
			err,
		)
	}

	return nil
}

// retryKey returns the Redis key used to store
// the number of attempts for one message.
func (w *NotificationWorker) retryKey(
	messageID string,
) string {
	return RetryKeyPrefix + messageID
}

// incrementAttempts increases the retry counter.
//
// Redis INCR returns int64. ++number
func (w *NotificationWorker) incrementAttempts(
	ctx context.Context,
	messageID string,
) (int64, error) {
	return w.redis.Incr(
		ctx,
		w.retryKey(messageID),
	).Result()
}

// clearAttempts removes the retry counter after
// successful processing or after sending to DLQ.
func (w *NotificationWorker) clearAttempts(
	ctx context.Context,
	messageID string,
) error {
	return w.redis.Del(
		ctx,
		w.retryKey(messageID),
	).Err()
}

// exponentialBackoff calculates:
//
// attempt 1 -> 1s
// attempt 2 -> 2s
// attempt 3 -> 4s
// attempt 4 -> 8s
//
// The value is capped at MaxBackoff.
func (w *NotificationWorker) exponentialBackoff(
	attempt int64,
) time.Duration {
	if attempt <= 0 {
		return w.initialBackoff
	}

	delay := w.initialBackoff

	for i := int64(1); i < attempt; i++ {
		if delay >= w.maxBackoff/2 {
			return w.maxBackoff
		}

		delay *= 2
	}

	if delay > w.maxBackoff {
		return w.maxBackoff
	}

	return delay
}

// backoffWithJitter adds +/-20% randomness.
//
// Example:
//
// 1s -> around 800ms-1200ms
// 2s -> around 1600ms-2400ms
func (w *NotificationWorker) backoffWithJitter(
	attempt int64,
) time.Duration {
	base := w.exponentialBackoff(attempt)

	// Random value between -0.2 and +0.2.
	jitter := (rand.Float64() * 0.4) - 0.2

	delay := float64(base) * (1 + jitter)

	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// waitBeforeRetry waits for the calculated backoff.
//
// The timer is context-aware, so shutdown/cancellation
// can interrupt it.
func (w *NotificationWorker) waitBeforeRetry(
	ctx context.Context,
	attempt int64,
) error {
	delay := w.backoffWithJitter(attempt)

	w.logger.Info(
		"waiting before retry",
		"attempt", attempt,
		"backoff", delay,
	)

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-timer.C:
		return nil
	}
}

// processMessage converts the Redis Stream event
// into a domain notification and stores it.
//
// EventID is unique in the database,
// which gives us idempotency.
func (w *NotificationWorker) processMessage(
	ctx context.Context,
	message redis.XMessage,
) error {
	rawEvent, ok := message.Values["event"].(string)
	if !ok {
		return fmt.Errorf(
			"event field missing from message %s",
			message.ID,
		)
	}

	var taskEvent event.TaskCreatedEvent

	if err := json.Unmarshal(
		[]byte(rawEvent),
		&taskEvent,
	); err != nil {
		return fmt.Errorf(
			"decode task event %s: %w",
			message.ID,
			err,
		)
	}

	notification := domain.Notification{
		EventID: taskEvent.EventID,
		TaskID:  taskEvent.TaskID,
		Message: fmt.Sprintf(
			"Task %d was created",
			taskEvent.TaskID,
		),
	}

	err := w.notificationRepo.Create(
		ctx,
		&notification,
	)
	if err != nil {
		// The same event may be delivered more than once.
		// Because EventID is unique, duplicates are safely ignored.
		if errors.Is(
			err,
			domain.ErrNotificationAlreadyExists,
		) {
			w.logger.Info(
				"duplicate notification ignored",
				"event_id", taskEvent.EventID,
				"message_id", message.ID,
			)

			return nil
		}

		return fmt.Errorf(
			"create notification: %w",
			err,
		)
	}

	return nil
}

// handleMessage processes exactly one Redis Stream message.
//
// Flow:
//
// 1. Create a timeout.
// 2. Process the event.
// 3. ACK on success.
// 4. On failure, increment attempts.
// 5. Retry if attempts < MaxAttempts.
// 6. Send to DLQ after MaxAttempts.
func (w *NotificationWorker) handleMessage(//إدارة محاولة واحدة + ACK/Retry/DLQ
	ctx context.Context,
	message redis.XMessage,
) error {
	start := time.Now()

	defer func() {
		observability.WorkerProcessingDuration.Observe(
			time.Since(start).Seconds(),
		)
	}()

	messageCtx, cancel := context.WithTimeout(
		ctx,
		MessageTimeout,
	)
	defer cancel()

	err := w.processMessage(
		messageCtx,
		message,
	)

	if err == nil {
		if ackErr := w.redis.XAck(
			ctx,
			w.stream,
			w.group,
			message.ID,
		).Err(); ackErr != nil {
			return fmt.Errorf(
				"ack message %s: %w",
				message.ID,
				ackErr,
			)
		}

		if clearErr := w.clearAttempts(
			ctx,
			message.ID,
		); clearErr != nil {
			w.logger.Warn(
				"failed to clear retry counter",
				"message_id", message.ID,
				"error", clearErr,
			)
		}

		observability.WorkerEventsProcessed.Inc()

		return nil
	}

	observability.WorkerErrors.Inc()

	attempts, attemptsErr := w.incrementAttempts(
		ctx,
		message.ID,
	)

	if attemptsErr != nil {
		return fmt.Errorf(
			"increment attempts for %s: %w",
			message.ID,
			attemptsErr,
		)
	}

	w.logger.Warn(
		"worker message failed",
		"message_id", message.ID,
		"attempts", attempts,
		"error", err,
	)

	// Retry while we still have attempts available.
	//
	// We intentionally DO NOT ACK the message.
	// Redis keeps it pending so XAUTOCLAIM can recover it.
	if attempts < int64(w.maxAttempts) {
		observability.WorkerRetries.Inc()

		return err
	}

	// Maximum attempts reached.
	//
	// Move the message to the DLQ.
	if dlqErr := w.sendToDLQ(
		ctx,
		message,
		attempts,
	); dlqErr != nil {
		return fmt.Errorf(
			"send message %s to DLQ: %w",
			message.ID,
			dlqErr,
		)
	}

	// ACK the original message only after
	// successfully copying it to the DLQ.
	if ackErr := w.redis.XAck(
		ctx,
		w.stream,
		w.group,
		message.ID,
	).Err(); ackErr != nil {
		return fmt.Errorf(
			"ack failed message %s: %w",
			message.ID,
			ackErr,
		)
	}

	if clearErr := w.clearAttempts(
		ctx,
		message.ID,
	); clearErr != nil {
		w.logger.Warn(
			"failed to clear retry counter after DLQ",
			"message_id", message.ID,
			"error", clearErr,
		)
	}

	observability.WorkerDLQ.Inc()

	return nil
}

// sendToDLQ copies the failed message
// to the dead-letter stream.

func (w *NotificationWorker) sendToDLQ(// = حفظ الرسالة التي فشلت نهائيًا
	ctx context.Context,
	message redis.XMessage,
	attempts int64,
) error {
	payload, err := json.Marshal(
		map[string]any{
			"event":    message.Values["event"],
			"attempts": attempts,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"marshal DLQ payload: %w",
			err,
		)
	}

	return w.redis.XAdd(
		ctx,
		&redis.XAddArgs{
			Stream: w.stream + ":dlq",
			Values: map[string]interface{}{
				"original_message_id": message.ID,
				"attempts":            attempts,
				"event":               string(payload),
			},
		},
	).Err()
}

// processBatch processes a bounded number
// of messages concurrently.
//
// Example:
//

// WorkerCount = 3
//
// At most 3 goroutines process messages concurrently.
func (w *NotificationWorker) processBatch( // BatchSize   = for msg 10 توزيع العمل على goroutines
	ctx context.Context,
	messages []redis.XMessage,
) error {
	if len(messages) == 0 {
		return nil
	}

	jobs := make(chan redis.XMessage)

	var wg sync.WaitGroup

	workerCount := w.workerCount

	if workerCount > len(messages) {
		workerCount = len(messages)
	}

	// Start bounded workers.
	for i := 0; i < workerCount; i++ {
		wg.Add(1)

		go func(workerID int) {
			defer wg.Done()

			for {
				select {
				case <-ctx.Done():
					return

				case message, ok := <-jobs:
					if !ok {
						return
					}

					if err := w.handleMessage(
						ctx,
						message,
					); err != nil {
						// A message-level failure is not a
						// batch-level failure.
						//
						// The message remains pending and
						// can be recovered by XAUTOCLAIM.
						w.logger.Warn(
							"message processing failed; message remains pending",
							"worker_id", workerID,
							"message_id", message.ID,
							"error", err,
						)
					}
				}
			}
		}(i + 1)
	}

	// Feed jobs to workers.
sendLoop:
	for _, message := range messages {
		select {
		case <-ctx.Done():
			break sendLoop

		case jobs <- message:
		}
	}

	close(jobs)

	wg.Wait()

	return nil
}

// processNewMessages reads new messages
// from the Redis Stream.
func (w *NotificationWorker) processNewMessages(//= جلب الرسائل الجديدة
	ctx context.Context,
) error {
	result, err := w.redis.XReadGroup(
		ctx,
		&redis.XReadGroupArgs{
			Group:    w.group,
			Consumer: w.consumer,
			Streams: []string{
				w.stream,
				">", // give me new msg that givent consumer yet
			},
			Count: int64(w.batchSize),
			Block: 2 * time.Second,
		},
	).Result()

	if err != nil {
		// if dont have msg
		if errors.Is(err, redis.Nil) {
			return nil
		}

		return fmt.Errorf(
			"read new messages: %w",
			err,
		)
	}

	for _, stream := range result {
		if err := w.processBatch(
			ctx,
			stream.Messages,
		); err != nil {
			return err
		}
	}

	return nil
}

// reclaimPendingMessages recovers messages
// that were delivered but never ack.
func (w *NotificationWorker) reclaimPendingMessages( //= استعادة الرسائل العالقة
	ctx context.Context,
) error {
	messages, _, err := w.redis.XAutoClaim(
		ctx,
		&redis.XAutoClaimArgs{
			Stream:   w.stream,
			Group:    w.group,
			Consumer: w.consumer,
			MinIdle:  w.claimIdle, //get pending msg after 10 seconds
			Start:    "0-0",       //  redis start search msg frm old point in stream
			Count:    int64(w.batchSize),
		},
	).Result()

	if err != nil {
		return fmt.Errorf(
			"auto claim pending messages: %w",
			err,
		)
	}

	return w.processBatch(
		ctx,
		messages,
	)
}

// Run starts the notification worker.
//
// The worker:
//
// 1. Ensures the consumer group exists.
// 2. Reclaims abandoned messages.
// 3. Reads new messages.
// 4. Stops cleanly when ctx is cancelled.
func (w *NotificationWorker) Run(// إدارة حياة الـ Worker
	ctx context.Context,
) error {
	if err := w.EnsureConsumerGroup(ctx); err != nil {
		return err
	}

	w.logger.Info(
		"notification worker started",
		"stream", w.stream,
		"group", w.group,
		"consumer", w.consumer,
		"workers", w.workerCount,
		"batch_size", w.batchSize,
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
			if err := w.reclaimPendingMessages(
				ctx,
			); err != nil {
				// Context cancellation during shutdown
				// is expected and should not be treated
				// as a real dependency failure.
				if !errors.Is(
					err,
					context.Canceled,
				) {
					w.logger.Error(
						"failed to reclaim pending messages",
						"error", err,
					)
				}
			}

		default:
			if err := w.processNewMessages(
				ctx,
			); err != nil {
				if errors.Is(
					err,
					context.Canceled,
				) {
					return ctx.Err()
				}

				w.logger.Error(
					"failed to process new messages",
					"error", err,
				)

				// Prevent a tight error loop when Redis
				// or another dependency is unavailable.
				timer := time.NewTimer(
					500 * time.Millisecond,
				) // give app latel time before retry

				select {
				case <-ctx.Done():
					timer.Stop()
					return ctx.Err()

				case <-timer.C:
				}
			}
		}
	}
}
