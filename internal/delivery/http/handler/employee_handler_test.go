package handler

import (
	"errors"
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

func (f *fakeEmployeeRepository) Create(employee domain.Employee) (domain.Employee, error) {
	f.createdEmployee = employee

	if f.createErr != nil {
		return domain.Employee{}, f.createErr
	}

	employee.ID = 1

	return employee, nil
}

func (f *fakeEmployeeRepository) GetByID(id int) (domain.Employee, error) {
	return domain.Employee{}, errors.New("not implemented")
}

func (f *fakeEmployeeRepository) GetAll(
	filter repository.EmployeeFilter,
	sortBy repository.EmployeeSort,
	offset, limit int,
) ([]domain.Employee, int, uint64, error) {
	return nil, 0, 0, errors.New("not implemented")
}

func (f *fakeEmployeeRepository) Update(employee domain.Employee) (domain.Employee, error) {
	return domain.Employee{}, errors.New("not implemented")
}

func (f *fakeEmployeeRepository) Delete(id int) error {
	return errors.New("not implemented")
}

func TestHealth(t *testing.T) {
	handler := &Handler{}

	req := httptest.NewRequest(
		"GET",
		"/health",
		nil,
	)

	rec := httptest.NewRecorder()

	handler.Health(rec, req)

	if rec.Code != 200 {
		t.Fatalf(
			"expected status 200, got %d",
			rec.Code,
		)
	}

	expectedBody := `{"status":"ok"}` + "\n"

	if rec.Body.String() != expectedBody {
		t.Fatalf(
			"expected body %q, got %q",
			expectedBody,
			rec.Body.String(),
		)
	}
}

func TestCreateEmployee_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		createErr      error
		expectedStatus int
	}{
		{
			name: "valid employee",
			body: `{
				"name": "Ali",
				"email": "ali@example.com",
				"department": "IT"
			}`,
			expectedStatus: 201,
		},
		{
			name: "repository error",
			body: `{
				"name": "Ali",
				"email": "ali@example.com",
				"department": "IT"
			}`,
			createErr:      errors.New("database error"),
			expectedStatus: 500,
		},
		{
			name: "empty name",
			body: `{
				"name": "",
				"email": "ali@example.com",
				"department": "IT"
			}`,
			expectedStatus: 400,
		},
		{
			name: "invalid JSON",
			body: `{
				"name": "Ali",
				"email":
			}`,
			expectedStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeEmployeeRepository{
				createErr: tt.createErr,
			}

			employeeUC := usecase.NewEmployeeUsecase(repo)

			handler := &Handler{
				usecase: employeeUC,
			}

			req := httptest.NewRequest(
				"POST",
				"/employees",
				strings.NewReader(tt.body),
			)

			req.Header.Set(
				"Content-Type",
				"application/json",
			)

			rec := httptest.NewRecorder()

			handler.CreateEmployee(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Fatalf(
					"expected status %d, got %d",
					tt.expectedStatus,
					rec.Code,
				)
			}
		})
	}
}
