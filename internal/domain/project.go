package domain

import "time"

type Project struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	OwnerID     int    `gorm:"not null;index" json:"owner_id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Version     uint   `gorm:"not null;default:1" json:"version"`
	//one to many relationship with tasks
	Tasks     []Task     `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
