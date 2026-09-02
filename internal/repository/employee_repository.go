package repository

import (
	"errors"

	"backend-api/internal/domain"
)

var ErrEmployeeNotFound = errors.New("employee not found")

type EmployeeRepository interface {
	Create(employee domain.Employee) (domain.Employee, error)
	GetByID(id int) (domain.Employee, error)
	GetAll(offset, limit int) ([]domain.Employee, int, uint64, error)
	Update(employee domain.Employee) (domain.Employee, error)
	Delete(id int) error
}
