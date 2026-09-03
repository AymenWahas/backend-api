package dto

type CreateTaskRequest struct {
	ProjectID uint   `json:"project_id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
}

type UpdateTaskRequest struct {
	ProjectID uint   `json:"project_id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
}
