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
	repo          repository.TaskRepository
	publisher     event.Publisher
	authorization *AuthorizationUsecase
}

func NewTaskUsecase(
	repo repository.TaskRepository,
	publisher event.Publisher,
	authorization *AuthorizationUsecase,
) *TaskUsecase {
	return &TaskUsecase{
		repo:          repo,
		publisher:     publisher,
		authorization: authorization,
	}
}

func (u *TaskUsecase) Create(
	ctx context.Context,
	userID int,
	task *domain.Task,
) error {

	allowed, err := u.authorization.CanModifyProject(
		ctx,
		userID,
		task.ProjectID,
	)

	if err != nil {
		return err
	}

	if !allowed {
		return domain.ErrForbidden
	}

	if err := u.repo.Create(ctx, task); err != nil {
		return err
	}

	taskEvent := event.TaskCreatedEvent{
		EventID:   uuid.NewString(),
		EventType: "task.created",
		TaskID:    task.ID,
		CreatedAt: time.Now().UTC(),
	}

	if err := u.publisher.PublishTaskCreated(
		ctx,
		taskEvent,
	); err != nil {
		return err
	}

	return nil
}

func (u *TaskUsecase) GetByID(
	ctx context.Context,
	userID int,
	id uint,
) (*domain.Task, error) {

	allowed, err := u.authorization.CanAccessTask(
		ctx,
		userID,
		id,
	)

	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, domain.ErrForbidden
	}

	return u.repo.GetByID(ctx, id)
}

func (u *TaskUsecase) GetAll(
	ctx context.Context,
	userID int,
) ([]domain.Task, error) {

	tasks, err := u.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Task, 0, len(tasks))

	for _, task := range tasks {
		allowed, err := u.authorization.CanAccessProject(
			ctx,
			userID,
			task.ProjectID,
		)

		if err != nil {
			return nil, err
		}

		if allowed {
			result = append(result, task)
		}
	}

	return result, nil
}

func (u *TaskUsecase) Update(
	ctx context.Context,
	userID int,
	task *domain.Task,
) error {

	allowed, err := u.authorization.CanModifyTask(
		ctx,
		userID,
		task.ID,
	)

	if err != nil {
		return err
	}

	if !allowed {
		return domain.ErrForbidden
	}

	return u.repo.Update(ctx, task)
}

func (u *TaskUsecase) Delete(
	ctx context.Context,
	userID int,
	id uint,
) error {

	allowed, err := u.authorization.CanDeleteTask(
		ctx,
		userID,
		id,
	)

	if err != nil {
		return err
	}

	if !allowed {
		return domain.ErrForbidden
	}

	return u.repo.Delete(ctx, id)
}
