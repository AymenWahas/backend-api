package domain

type Employee struct {
	ID         int     `gorm:"primaryKey" json:"id"`
	Name       string  `gorm:"not null" json:"name"`
	Email      string  `gorm:"unique;not null" json:"email"`
	Department *string `json:"department,omitempty"`
}
