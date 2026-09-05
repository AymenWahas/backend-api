package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"backend-api/internal/domain"
)

type TaskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{
		db: db,
	}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *TaskRepository) GetByID(ctx context.Context, id uint) (*domain.Task, error) {
	var task domain.Task

	err := r.db.WithContext(ctx).
		Preload("Project").
		First(&task, id).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrTaskNotFound
	}

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepository) GetAll(ctx context.Context) ([]domain.Task, error) {
	var tasks []domain.Task

	err := r.db.WithContext(ctx).
		Preload("Project").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Task{}).
		Where("id = ?", task.ID).
		Updates(task)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).
		Delete(&domain.Task{}, id)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}
