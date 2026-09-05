package usecase

import (
	"errors"
	"testing"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type fakeEmployeeRepository struct {
	//create
	createdEmployee domain.Employee
	createCalled    bool
	createErr       error
	//update
	updatedEmployee domain.Employee
	updateCalled    bool
	updateErr       error
	//getByID
	employeeByID  domain.Employee
	getByIDErr    error
	getByIDCalled bool
	//getAll
	employees    []domain.Employee
	total        int
	version      uint64
	getAllErr    error
	getAllCalled bool
	//delete
	deleteCalled bool
	deleteID     int
	deleteErr    error
}

func (f *fakeEmployeeRepository) Create(
	employee domain.Employee,
) (domain.Employee, error) {
	f.createCalled = true
	f.createdEmployee = employee

	if f.createErr != nil {
		return domain.Employee{}, f.createErr
	}

	return employee, nil
}

func (f *fakeEmployeeRepository) GetByID(
	id int,
) (domain.Employee, error) {
	f.getByIDCalled = true

	if f.getByIDErr != nil {
		return domain.Employee{}, f.getByIDErr
	}
	return f.employeeByID, nil
}

func (f *fakeEmployeeRepository) GetAll(
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
	employee domain.Employee,
) (domain.Employee, error) {

	f.updateCalled = true
	f.updatedEmployee = employee

	if f.updateErr != nil {
		return domain.Employee{}, f.updateErr
	}

	return employee, nil
}

func (f *fakeEmployeeRepository) Delete(id int) error {
	f.deleteCalled = true
	f.deleteID = id
	if f.deleteErr != nil {
		return f.deleteErr
	}
	return nil
}

func TestCreateEmployee(t *testing.T) {
	repositoryError := errors.New("repository error")

	tests := []struct {
		name          string
		employee      domain.Employee
		repositoryErr error
		wantName      string
		wantEmail     string
		wantErr       error
		wantCreate    bool
	}{
		{
			name: "valid employee trims spaces",
			employee: domain.Employee{
				Name:  "  Ali  ",
				Email: "  ali@example.com  ",
			},
			wantName:   "Ali",
			wantEmail:  "ali@example.com",
			wantCreate: true,
		},
		{
			name: "empty name",
			employee: domain.Employee{
				Name:  "   ",
				Email: "ali@example.com",
			},
			wantErr:    ErrInvalidEmployee,
			wantCreate: false,
		},
		{
			name: "empty email",
			employee: domain.Employee{
				Name:  "Ali",
				Email: "   ",
			},
			wantErr:    ErrInvalidEmployee,
			wantCreate: false,
		},
		{
			name: "repository error",
			employee: domain.Employee{
				Name:  "Ali",
				Email: "ali@example.com",
			},
			repositoryErr: repositoryError,
			wantErr:       repositoryError,
			wantCreate:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeEmployeeRepository{
				createErr: tt.repositoryErr,
			}

			uc := NewEmployeeUsecase(repo)

			got, err := uc.CreateEmployee(tt.employee)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}

			if tt.wantErr != nil {
				if repo.createCalled != tt.wantCreate {
					t.Fatalf(
						"expected Create called = %v, got %v",
						tt.wantCreate,
						repo.createCalled,
					)
				}

				return
			}

			if got.Name != tt.wantName {
				t.Errorf(
					"expected name %q, got %q",
					tt.wantName,
					got.Name,
				)
			}

			if got.Email != tt.wantEmail {
				t.Errorf(
					"expected email %q, got %q",
					tt.wantEmail,
					got.Email,
				)
			}

			if repo.createCalled != tt.wantCreate {
				t.Errorf(
					"expected Create called = %v, got %v",
					tt.wantCreate,
					repo.createCalled,
				)
			}
		})
	}
}
func TestUpdateEmployee(t *testing.T) {
	repositoryError := errors.New("repository error")

	tests := []struct {
		name          string
		employee      domain.Employee
		repositoryErr error
		wantName      string
		wantEmail     string
		wantErr       error
		wantUpdate    bool
	}{
		{
			name: "valid employee trims spaces",
			employee: domain.Employee{
				ID:    10,
				Name:  "  Ali  ",
				Email: "  ali@example.com  ",
			},
			wantName:   "Ali",
			wantEmail:  "ali@example.com",
			wantUpdate: true,
		},
		{
			name: "zero ID",
			employee: domain.Employee{
				ID:    0,
				Name:  "Ali",
				Email: "ali@example.com",
			},
			wantErr:    ErrInvalidEmployee,
			wantUpdate: false,
		},
		{
			name: "negative ID",
			employee: domain.Employee{
				ID:    -1,
				Name:  "Ali",
				Email: "ali@example.com",
			},
			wantErr:    ErrInvalidEmployee,
			wantUpdate: false,
		},
		{
			name: "empty name",
			employee: domain.Employee{
				ID:    10,
				Name:  "   ",
				Email: "ali@example.com",
			},
			wantErr:    ErrInvalidEmployee,
			wantUpdate: false,
		},
		{
			name: "empty email",
			employee: domain.Employee{
				ID:    10,
				Name:  "Ali",
				Email: "   ",
			},
			wantErr:    ErrInvalidEmployee,
			wantUpdate: false,
		},
		{
			name: "repository error",
			employee: domain.Employee{
				ID:    10,
				Name:  "Ali",
				Email: "ali@example.com",
			},
			repositoryErr: repositoryError,
			wantErr:       repositoryError,
			wantUpdate:    true,
		},
	}

	for _, tt := range tests {
		// giva all states anathor  name
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeEmployeeRepository{
				updateErr: tt.repositoryErr,
			}

			uc := NewEmployeeUsecase(repo)

			got, err := uc.UpdateEmployee(tt.employee)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}

			if repo.updateCalled != tt.wantUpdate {
				t.Fatalf(
					"expected Update called = %v, got %v",
					tt.wantUpdate,
					repo.updateCalled,
				)
			}

			if tt.wantErr != nil {
				return
			}

			if got.Name != tt.wantName {
				t.Errorf(
					"expected name %q, got %q",
					tt.wantName,
					got.Name,
				)
			}

			if got.Email != tt.wantEmail {
				t.Errorf(
					"expected email %q, got %q",
					tt.wantEmail,
					got.Email,
				)
			}

			if repo.updatedEmployee.Name != tt.wantName {
				t.Errorf(
					"expected repository name %q, got %q",
					tt.wantName,
					repo.updatedEmployee.Name,
				)
			}

			if repo.updatedEmployee.Email != tt.wantEmail {
				t.Errorf(
					"expected repository email %q, got %q",
					tt.wantEmail,
					repo.updatedEmployee.Email,
				)
			}
		})
	}
}
func TestGetEmployee(t *testing.T) {
	repositoryError := errors.New("repository error")

	tests := []struct {
		name          string
		id            int
		repositoryEmp domain.Employee
		repositoryErr error
		want          domain.Employee
		wantErr       error
		wantGetByID   bool
	}{
		{
			name: "employee found",
			id:   10,
			repositoryEmp: domain.Employee{
				ID:    10,
				Name:  "Ali",
				Email: "ali@example.com",
			},
			want: domain.Employee{
				ID:    10,
				Name:  "Ali",
				Email: "ali@example.com",
			},
			wantGetByID: true,
		},
		{
			name:          "repository error",
			id:            10,
			repositoryErr: repositoryError,
			wantErr:       repositoryError,
			wantGetByID:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeEmployeeRepository{
				employeeByID: tt.repositoryEmp,
				getByIDErr:   tt.repositoryErr,
			}

			uc := NewEmployeeUsecase(repo)

			got, err := uc.GetEmployee(tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}

			if repo.getByIDCalled != tt.wantGetByID {
				t.Fatalf(
					"expected GetByID called = %v, got %v",
					tt.wantGetByID,
					repo.getByIDCalled,
				)
			}

			if tt.wantErr != nil {
				return
			}

			if got != tt.want {
				t.Fatalf(
					"expected employee %+v, got %+v",
					tt.want,
					got,
				)
			}
		})
	}
}
func TestGetEmployees(t *testing.T) {
	repositoryError := errors.New("repository error")

	employees := []domain.Employee{
		{
			ID:    1,
			Name:  "Ali",
			Email: "ali@example.com",
		},
		{
			ID:    2,
			Name:  "Ahmed",
			Email: "ahmed@example.com",
		},
	}

	tests := []struct {
		name          string
		repositoryErr error
		wantLen       int
		wantTotal     int
		wantVersion   uint64
		wantErr       error
	}{
		{
			name:        "employees found",
			wantLen:     2,
			wantTotal:   2,
			wantVersion: 5,
		},
		{
			name:          "repository error",
			repositoryErr: repositoryError,
			wantErr:       repositoryError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeEmployeeRepository{
				employees: employees,
				total:     2,
				version:   5,
				getAllErr: tt.repositoryErr,
			}

			uc := NewEmployeeUsecase(repo)

			got, total, version, err := uc.GetEmployees(
				repository.EmployeeFilter{},
				repository.EmployeeSort{},
				0,
				10,
			)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}

			if !repo.getAllCalled {
				t.Fatal("expected GetAll to be called")
			}

			if tt.wantErr != nil {
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf(
					"expected length %d, got %d",
					tt.wantLen,
					len(got),
				)
			}

			if total != tt.wantTotal {
				t.Errorf(
					"expected total %d, got %d",
					tt.wantTotal,
					total,
				)
			}

			if version != tt.wantVersion {
				t.Errorf(
					"expected version %d, got %d",
					tt.wantVersion,
					version,
				)
			}
		})
	}
}
func TestDeleteEmployee(t *testing.T) {
	repositoryError := errors.New("employee not found")

	tests := []struct {
		name          string
		id            int
		repositoryErr error
		wantErr       error
	}{
		{
			name: "delete successfully",
			id:   10,
		},
		{
			name:          "repository error",
			id:            10,
			repositoryErr: repositoryError,
			wantErr:       repositoryError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeEmployeeRepository{
				deleteErr: tt.repositoryErr,
			}

			uc := NewEmployeeUsecase(repo)

			err := uc.DeleteEmployee(tt.id)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf(
					"expected error %v, got %v",
					tt.wantErr,
					err,
				)
			}

			if !repo.deleteCalled {
				t.Fatal("expected Delete to be called")
			}

			if repo.deleteID != tt.id {
				t.Fatalf(
					"expected delete ID %d, got %d",
					tt.id,
					repo.deleteID,
				)
			}
		})
	}
}

//go test ./internal/usecase -v -count=1

//go test ./internal/usecase -cover -count=1
