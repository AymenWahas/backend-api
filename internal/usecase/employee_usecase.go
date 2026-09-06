package usecase

import (
	"context"
	"errors"
	"strings"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

var ErrInvalidEmployee = errors.New("invalid employee")

type EmployeeUsecase struct {
	repo repository.EmployeeRepository
}

func NewEmployeeUsecase(repo repository.EmployeeRepository) *EmployeeUsecase {
	return &EmployeeUsecase{
		repo: repo,
	}
}

func (u *EmployeeUsecase) CreateEmployee(
	ctx context.Context,
	employee domain.Employee,
) (domain.Employee, error) {
	employee.Name = strings.TrimSpace(employee.Name)
	employee.Email = strings.TrimSpace(employee.Email)

	if employee.Name == "" || employee.Email == "" {
		return domain.Employee{}, ErrInvalidEmployee
	}

	return u.repo.Create(ctx, employee)
}

func (u *EmployeeUsecase) GetEmployee(ctx context.Context, id int) (domain.Employee, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *EmployeeUsecase) GetEmployees(ctx context.Context, filter repository.EmployeeFilter,
	sortBy repository.EmployeeSort, offset, limit int) (

	[]domain.Employee, int, uint64, error) {
	return u.repo.GetAll(ctx, filter, sortBy, offset, limit)
}

func (u *EmployeeUsecase) UpdateEmployee(ctx context.Context, employee domain.Employee) (domain.Employee, error) {
	employee.Name = strings.TrimSpace(employee.Name)
	employee.Email = strings.TrimSpace(employee.Email)

	if employee.ID <= 0 {
		return domain.Employee{}, ErrInvalidEmployee
	}

	if employee.Name == "" || employee.Email == "" {
		return domain.Employee{}, ErrInvalidEmployee
	}

	return u.repo.Update(ctx, employee)
}

func (u *EmployeeUsecase) DeleteEmployee(ctx context.Context, id int) error {
	return u.repo.Delete(ctx, id)
}
