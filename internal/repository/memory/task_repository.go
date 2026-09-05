package memory

import (
	"context"
	"sync"

	"backend-api/internal/domain"
)

type TaskRepository struct {
	mu     sync.RWMutex
	tasks  map[uint]domain.Task
	nextID uint
}

func NewTaskRepository() *TaskRepository {
	return &TaskRepository{
		tasks:  make(map[uint]domain.Task),
		nextID: 1,
	}
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	task.ID = r.nextID
	r.nextID++

	r.tasks[task.ID] = *task

	return nil
}

func (r *TaskRepository) GetByID(
	ctx context.Context,
	id uint,
) (*domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return nil, domain.ErrTaskNotFound
	}

	result := task

	return &result, nil
}

func (r *TaskRepository) GetAll(
	ctx context.Context,
) ([]domain.Task, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]domain.Task, 0, len(r.tasks))

	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *TaskRepository) Update(
	ctx context.Context,
	task *domain.Task,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[task.ID]; !ok {
		return domain.ErrTaskNotFound
	}

	r.tasks[task.ID] = *task

	return nil
}

func (r *TaskRepository) Delete(
	ctx context.Context,
	id uint,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return domain.ErrTaskNotFound
	}

	delete(r.tasks, id)

	return nil
}
