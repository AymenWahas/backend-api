package usecase

import (
	"context"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type ProjectTaskUsecase struct {
	projectRepo repository.ProjectRepository
	taskRepo    repository.TaskRepository
	txManager   TransactionManager
}

func NewProjectTaskUsecase(
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	txManager TransactionManager,
) *ProjectTaskUsecase {
	return &ProjectTaskUsecase{
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
		txManager:   txManager,
	}
}

func (u *ProjectTaskUsecase) CreateProjectWithTask(
	ctx context.Context,
	project *domain.Project,
	task *domain.Task,
) error {
	return u.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := u.projectRepo.Create(txCtx, project); err != nil {
			return err
		}

		task.ProjectID = project.ID

		if err := u.taskRepo.Create(txCtx, task); err != nil {
			return err
		}

		return nil
	})
}
