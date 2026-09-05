package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"backend-api/internal/domain"
)

type ProjectPostgresRepo struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectPostgresRepo {
	return &ProjectPostgresRepo{
		db: db,
	}
}

func (r *ProjectPostgresRepo) Create(ctx context.Context, project *domain.Project) error {
	result := r.db.WithContext(ctx).Create(project)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (r *ProjectPostgresRepo) GetByID(ctx context.Context, id uint) (*domain.Project, error) {
	var project domain.Project
	result := r.db.WithContext(ctx).First(&project, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrProjectNotFound
		}
		return nil, result.Error
	}
	return &project, nil
}

func (r *ProjectPostgresRepo) GetAll(ctx context.Context) ([]domain.Project, error) {
	var projects []domain.Project
	result := r.db.WithContext(ctx).Find(&projects)
	if result.Error != nil {
		return nil, result.Error
	}
	return projects, nil
}

func (r *ProjectPostgresRepo) Update(ctx context.Context, project *domain.Project) error {
	result := r.db.WithContext(ctx).Save(project)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}

func (r *ProjectPostgresRepo) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&domain.Project{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrProjectNotFound
	}
	return nil
}
