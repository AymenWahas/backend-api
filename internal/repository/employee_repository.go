package repository

import (
	"context"
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
	Create(ctx context.Context, employee domain.Employee) (domain.Employee, error)
	GetByID(ctx context.Context, id int) (domain.Employee, error)
	GetAll(
		ctx context.Context,
		filter EmployeeFilter,
		sort EmployeeSort,
		offset, limit int,
	) ([]domain.Employee, int, uint64, error)
	Update(ctx context.Context, employee domain.Employee) (domain.Employee, error)
	Delete(ctx context.Context, id int) error
}
