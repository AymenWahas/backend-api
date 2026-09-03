package domain

type Task struct {
	ID        uint   `json:"id"`
	ProjectID uint   `json:"project_id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
}
