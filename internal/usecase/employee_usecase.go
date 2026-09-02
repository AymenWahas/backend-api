package usecase

import (
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

func (u *EmployeeUsecase) CreateEmployee(employee domain.Employee) (domain.Employee, error) {
	employee.Name = strings.TrimSpace(employee.Name)
	employee.Email = strings.TrimSpace(employee.Email)

	if employee.Name == "" || employee.Email == "" {
		return domain.Employee{}, ErrInvalidEmployee
	}

	return u.repo.Create(employee)
}

func (u *EmployeeUsecase) GetEmployee(id int) (domain.Employee, error) {
	return u.repo.GetByID(id)
}

func (u *EmployeeUsecase) GetEmployees(offset, limit int) (

	[]domain.Employee, int, uint64, error) {
	return u.repo.GetAll(offset, limit)
}

func (u *EmployeeUsecase) UpdateEmployee(employee domain.Employee) (domain.Employee, error) {
	employee.Name = strings.TrimSpace(employee.Name)
	employee.Email = strings.TrimSpace(employee.Email)

	if employee.ID <= 0 {
		return domain.Employee{}, ErrInvalidEmployee
	}

	if employee.Name == "" || employee.Email == "" {
		return domain.Employee{}, ErrInvalidEmployee
	}

	return u.repo.Update(employee)
}

func (u *EmployeeUsecase) DeleteEmployee(id int) error {
	return u.repo.Delete(id)
}
