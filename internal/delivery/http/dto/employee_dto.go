package dto

import "strings"

type CreateEmployeeRequest struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Department *string `json:"department,omitempty"`
}

func (r CreateEmployeeRequest) Validate() bool {
	return strings.TrimSpace(r.Name) != "" &&
		strings.TrimSpace(r.Email) != ""
}

type UpdateEmployeeRequest struct {
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Department *string `json:"department,omitempty"`
}

func (r UpdateEmployeeRequest) Validate() bool {
	return strings.TrimSpace(r.Name) != "" &&
		strings.TrimSpace(r.Email) != ""
}

type EmployeeResponse struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Department *string `json:"department,omitempty"`
}
