package repository

import (
	"context"

	"backend-api/internal/domain"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByID(ctx context.Context, id uint) (*domain.Project, error)
	GetAll(ctx context.Context) ([]domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, id uint) error
}
