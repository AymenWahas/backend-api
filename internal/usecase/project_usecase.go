package usecase

import (
	"context"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type ProjectUsecase struct {
	repo repository.ProjectRepository
}

func NewProjectUsecase(repo repository.ProjectRepository) *ProjectUsecase {
	return &ProjectUsecase{
		repo: repo,
	}
}

func (u *ProjectUsecase) Create(
	ctx context.Context,
	project *domain.Project,
) error {
	return u.repo.Create(ctx, project)
}

func (u *ProjectUsecase) GetByID(
	ctx context.Context,
	id uint,
) (*domain.Project, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *ProjectUsecase) GetAll(
	ctx context.Context,
) ([]domain.Project, error) {
	return u.repo.GetAll(ctx)
}

func (u *ProjectUsecase) Update(
	ctx context.Context,
	project *domain.Project,
) error {
	return u.repo.Update(ctx, project)
}

func (u *ProjectUsecase) Delete(
	ctx context.Context,
	id uint,
) error {
	return u.repo.Delete(ctx, id)
}
