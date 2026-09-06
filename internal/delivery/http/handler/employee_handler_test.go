package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
	"backend-api/internal/usecase"
)

type fakeEmployeeRepository struct {
	createdEmployee domain.Employee
	createErr       error
}

func (f *fakeEmployeeRepository) Create(
	ctx context.Context,
	employee domain.Employee,
) (domain.Employee, error) {
	f.createdEmployee = employee

	if f.createErr != nil {
		return domain.Employee{}, f.createErr
	}

	employee.ID = 1

	return employee, nil
}

func (f *fakeEmployeeRepository) GetByID(
	ctx context.Context,
	id int,
) (domain.Employee, error) {
	return domain.Employee{}, errors.New("not implemented")
}

func (f *fakeEmployeeRepository) GetAll(
	ctx context.Context,
	filter repository.EmployeeFilter,
	sortBy repository.EmployeeSort,
	offset, limit int,
) ([]domain.Employee, int, uint64, error) {
	return nil, 0, 0, errors.New("not implemented")
}

func (f *fakeEmployeeRepository) Update(
	ctx context.Context,
	employee domain.Employee,
) (domain.Employee, error) {
	return domain.Employee{}, errors.New("not implemented")
}

func (f *fakeEmployeeRepository) Delete(
	ctx context.Context,
	id int,
) error {
	return errors.New("not implemented")
}

func TestCreateEmployee(t *testing.T) {
	repo := &fakeEmployeeRepository{}
	employeeUC := usecase.NewEmployeeUsecase(repo)
	projectUC := usecase.NewProjectUsecase(nil)
	taskUC := usecase.NewTaskUsecase(nil)

	h := NewHandler(employeeUC, projectUC, taskUC)
	body := `{
		"name": "Ali",
		"email": "ali@example.com"
	}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/employees",
		strings.NewReader(body),
	)

	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	h.CreateEmployee(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusCreated,
			rec.Code,
		)
	}

	if repo.createdEmployee.Name != "Ali" {
		t.Errorf(
			"expected name %q, got %q",
			"Ali",
			repo.createdEmployee.Name,
		)
	}

	if repo.createdEmployee.Email != "ali@example.com" {
		t.Errorf(
			"expected email %q, got %q",
			"ali@example.com",
			repo.createdEmployee.Email,
		)
	}
}
