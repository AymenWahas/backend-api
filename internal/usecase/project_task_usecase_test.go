package usecase

import (
	"context"
	"errors"
	"testing"

	"backend-api/internal/domain"
)

type fakeProjectRepository struct {
	created *domain.Project
}

func (r *fakeProjectRepository) Create(
	ctx context.Context,
	project *domain.Project,
) error {
	project.ID = 100
	r.created = project
	return nil
}

func (r *fakeProjectRepository) GetByID(
	ctx context.Context,
	id uint,
) (*domain.Project, error) {
	return nil, nil
}

func (r *fakeProjectRepository) GetAll(
	ctx context.Context,
) ([]domain.Project, error) {
	return nil, nil
}

func (r *fakeProjectRepository) Update(
	ctx context.Context,
	project *domain.Project,
) error {
	return nil
}

func (r *fakeProjectRepository) Delete(
	ctx context.Context,
	id uint,
) error {
	return nil
}

type fakeTaskRepository struct {
	err error
}

func (r *fakeTaskRepository) Create(
	ctx context.Context,
	task *domain.Task,
) error {
	return r.err
}

func (r *fakeTaskRepository) GetByID(
	ctx context.Context,
	id uint,
) (*domain.Task, error) {
	return nil, nil
}

func (r *fakeTaskRepository) GetAll(
	ctx context.Context,
) ([]domain.Task, error) {
	return nil, nil
}

func (r *fakeTaskRepository) Update(
	ctx context.Context,
	task *domain.Task,
) error {
	return nil
}

func (r *fakeTaskRepository) Delete(
	ctx context.Context,
	id uint,
) error {
	return nil
}

type fakeTransactionManager struct {
	called bool
}

func (tm *fakeTransactionManager) WithinTransaction(
	ctx context.Context,
	fn func(ctx context.Context) error,
) error {
	tm.called = true
	return fn(ctx)
}

func TestCreateProjectWithTaskRollback(t *testing.T) {
	projectRepo := &fakeProjectRepository{}

	taskRepo := &fakeTaskRepository{
		err: errors.New("task creation failed"),
	}

	txManager := &fakeTransactionManager{}

	uc := NewProjectTaskUsecase(
		projectRepo,
		taskRepo,
		txManager,
	)

	project := &domain.Project{
		Name: "Rollback Project",
	}

	task := &domain.Task{
		Title: "Rollback Task",
	}

	err := uc.CreateProjectWithTask(
		context.Background(),
		project,
		task,
	)

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !txManager.called {
		t.Fatal("expected transaction manager to be called")
	}

	if projectRepo.created == nil {
		t.Fatal("expected project repository Create to be called")
	}

	if task.ProjectID != project.ID {
		t.Fatalf(
			"expected task project ID %d, got %d",
			project.ID,
			task.ProjectID,
		)
	}
}
