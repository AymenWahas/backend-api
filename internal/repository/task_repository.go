package repository

import (
	"context"

	"backend-api/internal/domain"
)

type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	GetByID(ctx context.Context, id uint) (*domain.Task, error)
	GetAll(ctx context.Context) ([]domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id uint) error
}
