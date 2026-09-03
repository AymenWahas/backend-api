package dto

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type EmployeeListResponse struct {
	Data       []EmployeeResponse `json:"data"`
	Pagination Pagination         `json:"pagination"`
}
