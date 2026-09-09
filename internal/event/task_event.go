package event

import "time"

type TaskCreatedEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	TaskID    uint      `json:"task_id"`
	CreatedAt time.Time `json:"created_at"`
}
