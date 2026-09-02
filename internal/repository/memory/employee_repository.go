package memory

import (
	"sort"
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
	offset, limit int,
) ([]domain.Employee, int, uint64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	employees := make([]domain.Employee, 0, len(r.employees))

	for _, employee := range r.employees {
		employees = append(employees, employee)
	}
	sort.Slice(employees, func(i, j int) bool {
		return employees[i].ID < employees[j].ID
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
