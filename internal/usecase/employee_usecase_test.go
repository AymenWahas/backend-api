package usecase

import (
	"context"
	"errors"
	"testing"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type fakeEmployeeRepository struct {
	createdEmployee domain.Employee
	createCalled    bool
	createErr       error

	updatedEmployee domain.Employee
	updateCalled    bool
	updateErr       error

	employeeByID  domain.Employee
	getByIDErr    error
	getByIDCalled bool

	employees    []domain.Employee
	total        int
	version      uint64
	getAllErr    error
	getAllCalled bool

	deleteCalled bool
	deleteID     int
	deleteErr    error

	receivedContext context.Context
}

func (f *fakeEmployeeRepository) Create(
	ctx context.Context,
	employee domain.Employee,
) (domain.Employee, error) {
	f.createCalled = true
	f.createdEmployee = employee
	f.receivedContext = ctx

	if f.createErr != nil {
		return domain.Employee{}, f.createErr
	}

	return employee, nil
}

func (f *fakeEmployeeRepository) GetByID(
	ctx context.Context,
	id int,
) (domain.Employee, error) {
	f.getByIDCalled = true

	if f.getByIDErr != nil {
		return domain.Employee{}, f.getByIDErr
	}

	return f.employeeByID, nil
}

func (f *fakeEmployeeRepository) GetAll(
	ctx context.Context,
	filter repository.EmployeeFilter,
	sortBy repository.EmployeeSort,
	offset, limit int,
) ([]domain.Employee, int, uint64, error) {
	f.getAllCalled = true

	if f.getAllErr != nil {
		return nil, 0, 0, f.getAllErr
	}

	return f.employees, f.total, f.version, nil
}

func (f *fakeEmployeeRepository) Update(
	ctx context.Context,
	employee domain.Employee,
) (domain.Employee, error) {
	f.updateCalled = true
	f.updatedEmployee = employee

	if f.updateErr != nil {
		return domain.Employee{}, f.updateErr
	}

	return employee, nil
}

func (f *fakeEmployeeRepository) Delete(
	ctx context.Context,
	id int,
) error {
	f.deleteCalled = true
	f.deleteID = id

	if f.deleteErr != nil {
		return f.deleteErr
	}

	return nil
}

func TestCreateEmployee(t *testing.T) {
	repo := &fakeEmployeeRepository{}
	uc := NewEmployeeUsecase(repo)

	employee := domain.Employee{
		Name:  "  Ali  ",
		Email: "  ali@example.com  ",
	}

	got, err := uc.CreateEmployee(context.Background(), employee)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.createCalled {
		t.Fatal("expected Create to be called")
	}

	if repo.createdEmployee.Name != "Ali" {
		t.Errorf("expected trimmed name %q, got %q",
			"Ali", repo.createdEmployee.Name)
	}

	if repo.createdEmployee.Email != "ali@example.com" {
		t.Errorf("expected trimmed email %q, got %q",
			"ali@example.com", repo.createdEmployee.Email)
	}

	if got.Name != "Ali" {
		t.Errorf("expected name %q, got %q", "Ali", got.Name)
	}
}

func TestCreateEmployeeInvalid(t *testing.T) {
	repo := &fakeEmployeeRepository{}
	uc := NewEmployeeUsecase(repo)

	employee := domain.Employee{
		Name:  "   ",
		Email: "ali@example.com",
	}

	_, err := uc.CreateEmployee(context.Background(), employee)

	if !errors.Is(err, ErrInvalidEmployee) {
		t.Fatalf("expected ErrInvalidEmployee, got %v", err)
	}

	if repo.createCalled {
		t.Fatal("repository Create should not be called")
	}
}

func TestGetEmployee(t *testing.T) {
	repo := &fakeEmployeeRepository{
		employeeByID: domain.Employee{
			ID:    10,
			Name:  "Ali",
			Email: "ali@example.com",
		},
	}

	uc := NewEmployeeUsecase(repo)

	got, err := uc.GetEmployee(context.Background(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.getByIDCalled {
		t.Fatal("expected GetByID to be called")
	}

	if got.ID != 10 {
		t.Errorf("expected ID 10, got %d", got.ID)
	}
}

func TestGetEmployees(t *testing.T) {
	repo := &fakeEmployeeRepository{
		employees: []domain.Employee{
			{
				ID:    1,
				Name:  "Ali",
				Email: "ali@example.com",
			},
		},
		total:   1,
		version: 5,
	}

	uc := NewEmployeeUsecase(repo)

	filter := repository.EmployeeFilter{
		Name: "Ali",
	}

	sortBy := repository.EmployeeSort{
		Field: "name",
		Desc:  false,
	}

	got, total, version, err :=
		uc.GetEmployees(context.Background(), filter, sortBy, 0, 10)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.getAllCalled {
		t.Fatal("expected GetAll to be called")
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 employee, got %d", len(got))
	}

	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}

	if version != 5 {
		t.Errorf("expected version 5, got %d", version)
	}
}

func TestUpdateEmployee(t *testing.T) {
	repo := &fakeEmployeeRepository{}
	uc := NewEmployeeUsecase(repo)

	employee := domain.Employee{
		ID:    10,
		Name:  "  Ali  ",
		Email: "  ali@example.com  ",
	}

	got, err := uc.UpdateEmployee(
		context.Background(),
		employee,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.updateCalled {
		t.Fatal("expected Update to be called")
	}

	if repo.updatedEmployee.Name != "Ali" {
		t.Errorf("expected trimmed name %q, got %q",
			"Ali", repo.updatedEmployee.Name)
	}

	if repo.updatedEmployee.Email != "ali@example.com" {
		t.Errorf("expected trimmed email %q, got %q",
			"ali@example.com", repo.updatedEmployee.Email)
	}

	if got.ID != 10 {
		t.Errorf("expected ID 10, got %d", got.ID)
	}
}

func TestUpdateEmployeeInvalidID(t *testing.T) {
	repo := &fakeEmployeeRepository{}
	uc := NewEmployeeUsecase(repo)

	employee := domain.Employee{
		ID:    0,
		Name:  "Ali",
		Email: "ali@example.com",
	}

	_, err := uc.UpdateEmployee(
		context.Background(),
		employee,
	)

	if !errors.Is(err, ErrInvalidEmployee) {
		t.Fatalf("expected ErrInvalidEmployee, got %v", err)
	}

	if repo.updateCalled {
		t.Fatal("repository Update should not be called")
	}
}

func TestDeleteEmployee(t *testing.T) {
	repo := &fakeEmployeeRepository{}
	uc := NewEmployeeUsecase(repo)

	err := uc.DeleteEmployee(
		context.Background(),
		10,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !repo.deleteCalled {
		t.Fatal("expected Delete to be called")
	}

	if repo.deleteID != 10 {
		t.Errorf("expected delete ID 10, got %d", repo.deleteID)
	}
}

func TestCreateEmployeeContextPropagation(t *testing.T) {
	repo := &fakeEmployeeRepository{}
	uc := NewEmployeeUsecase(repo)

	ctx := context.WithValue(
		context.Background(),
		"request-id",
		"test-123",
	)

	employee := domain.Employee{
		Name:  "Ali",
		Email: "ali@example.com",
	}

	_, err := uc.CreateEmployee(ctx, employee)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.receivedContext != ctx {
		t.Fatal("expected the same context to reach the repository")
	}
}
