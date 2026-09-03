package memory

import (
	"context"
	"errors"
	"sync"

	"backend-api/internal/domain"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[uint]domain.Project
	nextID   uint
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		projects: make(map[uint]domain.Project),
		nextID:   1,
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	project.ID = r.nextID
	r.nextID++

	r.projects[project.ID] = *project

	return nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id uint) (*domain.Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	project, ok := r.projects[id]
	if !ok {
		return nil, ErrProjectNotFound
	}

	result := project
	return &result, nil
}

func (r *ProjectRepository) GetAll(ctx context.Context) ([]domain.Project, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]domain.Project, 0, len(r.projects))

	for _, project := range r.projects {
		projects = append(projects, project)
	}

	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.projects[project.ID]; !ok {
		return ErrProjectNotFound
	}

	r.projects[project.ID] = *project

	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id uint) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.projects[id]; !ok {
		return ErrProjectNotFound
	}

	delete(r.projects, id)

	return nil
}
