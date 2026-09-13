package domain

import "time"

type User struct {
	ID           int       `json:"id"`
	EmployeeID   int       `json:"employee_id"`
	PasswordHash string    `json:"-"` // This field will not be serialized to JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
