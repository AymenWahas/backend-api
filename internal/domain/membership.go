package domain

import "time"

type Membership struct {
	EmployeeID int       `gorm:"primaryKey;index" json:"employee_id"`
	ProjectID  uint      `gorm:"primaryKey;index" json:"project_id"`
	Role       string    `gorm:"not null;default:'member'" json:"role"`
	IsActive   bool      `gorm:"not null;default:true" json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}