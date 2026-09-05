package domain

// حذف السطر التالي:
// var ErrProjectNotFound = errors.New("project not found")

type Project struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"not null" json:"name"`
	Description string `json:"description"`
	Tasks       []Task `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
}
