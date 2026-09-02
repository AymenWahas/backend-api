package repository

import (
	"errors"

	"backend-api/internal/domain"
)

var ErrEmployeeNotFound = errors.New("employee not found")

type EmployeeFilter struct {
	Name  string
	Email string
}
type EmployeeSort struct {
	Field string
	Desc  bool
}

type EmployeeRepository interface {
	Create(employee domain.Employee) (domain.Employee, error)
	GetByID(id int) (domain.Employee, error)
	GetAll(
		filter EmployeeFilter,
		sort EmployeeSort,
		offset, limit int,
	) ([]domain.Employee, int, uint64, error)
	Update(employee domain.Employee) (domain.Employee, error)
	Delete(id int) error
}
