package memory

import (
	"sort"
	"strings"
	"sync"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type EmployeeRepository struct {
	mu        sync.RWMutex
	employees map[int]domain.Employee
	nextID    int
	version   uint64
}

func NewEmployeeRepository() *EmployeeRepository {
	return &EmployeeRepository{
		employees: make(map[int]domain.Employee),
		nextID:    1,
	}
}

func (r *EmployeeRepository) Create(employee domain.Employee) (domain.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	employee.ID = r.nextID
	r.nextID++

	r.employees[employee.ID] = employee
	r.version++
	return employee, nil
}

func (r *EmployeeRepository) GetByID(id int) (domain.Employee, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employee, ok := r.employees[id]
	if !ok {
		return domain.Employee{}, repository.ErrEmployeeNotFound
	}

	return employee, nil
}
func (r *EmployeeRepository) GetAll(
	filter repository.EmployeeFilter,
	sortBy repository.EmployeeSort,
	offset, limit int,
) ([]domain.Employee, int, uint64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employees := make([]domain.Employee, 0, len(r.employees))

	for _, employee := range r.employees {

		if filter.Name != "" &&
			!strings.Contains(
				strings.ToLower(employee.Name),
				strings.ToLower(filter.Name),
			) {
			continue
		}

		if filter.Email != "" &&
			!strings.Contains(
				strings.ToLower(employee.Email),
				strings.ToLower(filter.Email),
			) {
			continue
		}

		employees = append(employees, employee)
	}
	sort.Slice(employees, func(i, j int) bool {
		var less bool
		var greater bool

		switch sortBy.Field {
		case "name":
			left := strings.ToLower(employees[i].Name)
			right := strings.ToLower(employees[j].Name)

			less = left < right
			greater = left > right

		case "email":
			left := strings.ToLower(employees[i].Email)
			right := strings.ToLower(employees[j].Email)

			less = left < right
			greater = left > right

		default:
			less = employees[i].ID < employees[j].ID
			greater = employees[i].ID > employees[j].ID
		}

		if sortBy.Desc {
			return greater
		}

		return less
	})
	total := len(employees)

	if offset >= total {
		return []domain.Employee{}, total, r.version, nil
	}

	end := offset + limit

	if end > total {
		end = total
	}

	return employees[offset:end], total, r.version, nil
}
func (r *EmployeeRepository) Update(employee domain.Employee) (domain.Employee, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.employees[employee.ID]; !ok {
		return domain.Employee{}, repository.ErrEmployeeNotFound
	}

	r.employees[employee.ID] = employee
	r.version++
	return employee, nil
}

func (r *EmployeeRepository) Delete(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.employees[id]; !ok {
		return repository.ErrEmployeeNotFound
	}

	delete(r.employees, id)
	r.version++
	return nil
}
