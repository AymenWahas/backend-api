package usecase

import (
	"context"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type TaskUsecase struct {
	repo repository.TaskRepository
}

func NewTaskUsecase(repo repository.TaskRepository) *TaskUsecase {
	return &TaskUsecase{
		repo: repo,
	}
}

func (u *TaskUsecase) Create(
	ctx context.Context,
	task *domain.Task,
) error {
	return u.repo.Create(ctx, task)
}

func (u *TaskUsecase) GetByID(
	ctx context.Context,
	id uint,
) (*domain.Task, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *TaskUsecase) GetAll(
	ctx context.Context,
) ([]domain.Task, error) {
	return u.repo.GetAll(ctx)
}

func (u *TaskUsecase) Update(
	ctx context.Context,
	task *domain.Task,
) error {
	return u.repo.Update(ctx, task)
}

func (u *TaskUsecase) Delete(
	ctx context.Context,
	id uint,
) error {
	return u.repo.Delete(ctx, id)
}
