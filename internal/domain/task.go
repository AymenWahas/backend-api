package domain

import "time"

type Task struct {
	ID        uint       `gorm:"primaryKey" json:"id"`
	//index is used to improve the performance of queries that filter or sort by this field

	ProjectID uint       `gorm:"not null;index" json:"project_id"`
	Title     string     `gorm:"not null" json:"title"`
	Status    string     `gorm:"not null;default:'pending'" json:"status"`
	// containt information about the project to which the task belongs
	
	Project   Project    `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
}
