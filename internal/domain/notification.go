package domain

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	EventID   string    `gorm:"not null;uniqueIndex" json:"event_id"`
	TaskID    uint      `gorm:"not null" json:"task_id"`
	Message   string    `gorm:"not null" json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
