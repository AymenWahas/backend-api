package usecase

import (
	"testing"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type fakeEmployeeRepository struct {
	createdEmployee domain.Employee
	createCalled    bool
}

func (f *fakeEmployeeRepository) Create(
	employee domain.Employee,
) (domain.Employee, error) {

	f.createCalled = true
	f.createdEmployee = employee

	return employee, nil
}

func (f *fakeEmployeeRepository) GetByID(
	id int,
) (domain.Employee, error) {
	return domain.Employee{}, nil
}

func (f *fakeEmployeeRepository) GetAll(
	filter repository.EmployeeFilter,
	sortBy repository.EmployeeSort,
	offset, limit int,
) ([]domain.Employee, int, uint64, error) {
	return nil, 0, 0, nil
}

func (f *fakeEmployeeRepository) Update(
	employee domain.Employee,
) (domain.Employee, error) {
	return employee, nil
}

func (f *fakeEmployeeRepository) Delete(id int) error {
	return nil
}

func TestCreateEmployee(t *testing.T) {
	repo := &fakeEmployeeRepository{}

	uc := NewEmployeeUsecase(repo)

	employee := domain.Employee{
		Name:  "  Ali  ",
		Email: "  ali@example.com  ",
	}

	got, err := uc.CreateEmployee(employee)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Name != "Ali" {
		t.Errorf("expected name %q, got %q", "Ali", got.Name)
	}

	if got.Email != "ali@example.com" {
		t.Errorf(
			"expected email %q, got %q",
			"ali@example.com",
			got.Email,
		)
	}

	if !repo.createCalled {
		t.Error("expected repository Create to be called")
	}
}
