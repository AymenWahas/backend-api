package domain

type Task struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	ProjectID uint   `gorm:"not null;index" json:"project_id"`
	Title     string `gorm:"not null" json:"title"`
	Status    string `gorm:"default:'pending'" json:"status"`
	// العلاقة: Task ينتمي إلى Project
	Project Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}
