package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"backend-api/internal/domain"
	"backend-api/internal/event"
	"backend-api/internal/repository"
)

type TaskUsecase struct {
	repo      repository.TaskRepository
	publisher event.Publisher
}

func NewTaskUsecase(
	repo repository.TaskRepository,
	publisher event.Publisher,
) *TaskUsecase {
	return &TaskUsecase{
		repo:      repo,
		publisher: publisher,
	}
}

func (u *TaskUsecase) Create(
	ctx context.Context,
	task *domain.Task,
) error {
	if err := u.repo.Create(ctx, task); err != nil {
		return err
	}

	taskEvent := event.TaskCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "task.created",
		TaskID:    task.ID,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.publisher.PublishTaskCreated(ctx, taskEvent); err != nil {
		return err
	}

	return nil
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
