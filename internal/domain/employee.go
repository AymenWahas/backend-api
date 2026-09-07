package domain

import "time"

type Employee struct {
	ID         int        `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"not null" json:"name"`
	Email      string     `gorm:"unique;not null" json:"email"`
	Department *string    `json:"department,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}




