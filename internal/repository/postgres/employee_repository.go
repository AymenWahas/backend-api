package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type EmployeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) *EmployeeRepository {
	return &EmployeeRepository{
		db: db,
	}
}

// Create creates a new employee in PostgreSQL.
func (r *EmployeeRepository) Create(
	ctx context.Context,
	employee domain.Employee,
) (domain.Employee, error) {
	result := r.db.WithContext(ctx).Create(&employee)

	if result.Error != nil {
		return domain.Employee{}, result.Error
	}

	return employee, nil
}

// GetByID returns one employee by ID.
func (r *EmployeeRepository) GetByID(
	ctx context.Context,
	id int,
) (domain.Employee, error) {
	var employee domain.Employee

	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&employee)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.Employee{}, domain.ErrEmployeeNotFound
	}

	if result.Error != nil {
		return domain.Employee{}, result.Error
	}

	return employee, nil
}

// GetAll returns employees with filtering, sorting and pagination.
func (r *EmployeeRepository) GetAll(
	ctx context.Context,
	filter repository.EmployeeFilter,
	sortBy repository.EmployeeSort,
	offset, limit int,
) ([]domain.Employee, int, uint64, error) {
	query := r.db.WithContext(ctx).
		Model(&domain.Employee{})

	// Filtering
	if filter.Name != "" {
		query = query.Where(
			"LOWER(name) LIKE LOWER(?)",
			"%"+filter.Name+"%",
		)
	}

	if filter.Email != "" {
		query = query.Where(
			"LOWER(email) LIKE LOWER(?)",
			"%"+filter.Email+"%",
		)
	}

	// Total number of employees after filtering.
	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, 0, err
	}

	// Sorting
	order := "id ASC"

	switch sortBy.Field {
	case "name":
		order = "name ASC"
	case "email":
		order = "email ASC"
	default:
		order = "id ASC"
	}

	if sortBy.Desc {
		order = order[:len(order)-3] + " DESC"
	}

	query = query.Order(order)

	// Pagination
	query = query.
		Offset(offset).
		Limit(limit)

	var employees []domain.Employee

	if err := query.Find(&employees).Error; err != nil {
		return nil, 0, 0, err
	}

	return employees, int(total), 0, nil
}

// Update updates an existing employee.
func (r *EmployeeRepository) Update(
	ctx context.Context,
	employee domain.Employee,
) (domain.Employee, error) {
	result := r.db.WithContext(ctx).
		Model(&domain.Employee{}).
		Where("id = ?", employee.ID).
		Updates(map[string]interface{}{
			"name":       employee.Name,
			"email":      employee.Email,
			"department": employee.Department,
		})

	if result.Error != nil {
		return domain.Employee{}, result.Error
	}

	if result.RowsAffected == 0 {
		return domain.Employee{}, domain.ErrEmployeeNotFound
	}

	return r.GetByID(ctx, employee.ID)
}

// Delete deletes an employee by ID.
func (r *EmployeeRepository) Delete(
	ctx context.Context,
	id int,
) error {
	result := r.db.WithContext(ctx).
		Delete(&domain.Employee{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrEmployeeNotFound
	}

	return nil
}
