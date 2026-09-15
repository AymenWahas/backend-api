package usecase

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"backend-api/internal/cache"
	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

const projectCacheTTL = 5 * time.Minute

type ProjectUsecase struct {
	repo  repository.ProjectRepository
	cache *cache.RedisClient
	auth  *AuthorizationUsecase
}

func NewProjectUsecase(
	repo repository.ProjectRepository,
	cache *cache.RedisClient,
	auth *AuthorizationUsecase,
) *ProjectUsecase {
	return &ProjectUsecase{
		repo:  repo,
		cache: cache,
		auth:  auth,
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

	key := fmt.Sprintf("project:%d", id)

	var cachedProject domain.Project

	err := u.cache.Get(
		ctx,
		key,
		&cachedProject,
	)

	if err == nil {
		slog.Info(
			"project cache hit",
			"project_id",
			id,
			"cache_key",
			key,
		)

		return &cachedProject, nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		slog.Warn(
			"project cache unavailable",
			"project_id",
			id,
			"cache_key",
			key,
			"error",
			err,
		)
	} else {
		slog.Info(
			"project cache miss",
			"project_id",
			id,
			"cache_key",
			key,
		)
	}

	project, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := u.cache.Set(
		ctx,
		key,
		project,
		projectCacheTTL,
	); err != nil {
		slog.Warn(
			"project cache set failed",
			"project_id",
			id,
			"cache_key",
			key,
			"error",
			err,
		)

		return project, nil
	}

	slog.Info(
		"project cached",
		"project_id",
		id,
		"cache_key",
		key,
		"ttl",
		projectCacheTTL.String(),
	)

	return project, nil
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

	if err := u.repo.Update(ctx, project); err != nil {
		slog.Error(
			"project update failed",
			"project_id",
			project.ID,
			"version",
			project.Version,
			"error",
			err,
		)

		return err
	}

	key := fmt.Sprintf(
		"project:%d",
		project.ID,
	)

	if err := u.cache.Delete(ctx, key); err != nil {
		slog.Warn(
			"project cache invalidation failed",
			"project_id",
			project.ID,
			"cache_key",
			key,
			"error",
			err,
		)

		return nil
	}

	slog.Info(
		"project cache invalidated",
		"project_id",
		project.ID,
		"cache_key",
		key,
	)

	return nil
}

func (u *ProjectUsecase) Delete(
	ctx context.Context,
	id uint,
) error {

	if err := u.repo.Delete(ctx, id); err != nil {
		return err
	}

	key := fmt.Sprintf(
		"project:%d",
		id,
	)

	if err := u.cache.Delete(ctx, key); err != nil {
		slog.Warn(
			"project cache invalidation failed",
			"project_id",
			id,
			"cache_key",
			key,
			"error",
			err,
		)

		return nil
	}

	slog.Info(
		"project cache invalidated",
		"project_id",
		id,
		"cache_key",
		key,
	)

	return nil
}

func (u *ProjectUsecase) Authorization() *AuthorizationUsecase {
	return u.auth
}
func (u *ProjectUsecase) GetAuthorized(
	ctx context.Context,
	userID int,
	projectID uint,
) (*domain.Project, error) {

	project, err := u.GetByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	allowed, err := u.auth.CanAccessProject(
		ctx,
		userID,
		projectID,
	)

	if err != nil {
		return nil, err
	}

	if !allowed {
		return nil, domain.ErrForbidden
	}

	return project, nil
}

func (u *ProjectUsecase) GetAllAuthorized(
	ctx context.Context,
	userID int,
) ([]domain.Project, error) {

	projects, err := u.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Project, 0, len(projects))

	for _, project := range projects {
		allowed, err := u.auth.CanAccessProject(
			ctx,
			userID,
			project.ID,
		)

		if err != nil {
			return nil, err
		}

		if allowed {
			result = append(result, project)
		}
	}

	return result, nil
}

func (u *ProjectUsecase) UpdateAuthorized(
	ctx context.Context,
	userID int,
	project *domain.Project,
) error {

	allowed, err := u.auth.CanModifyProject(
		ctx,
		userID,
		project.ID,
	)

	if err != nil {
		return err
	}

	if !allowed {
		return domain.ErrForbidden
	}

	return u.Update(ctx, project)
}

func (u *ProjectUsecase) DeleteAuthorized(
	ctx context.Context,
	userID int,
	projectID uint,
) error {

	allowed, err := u.auth.CanDeleteProject(
		ctx,
		userID,
		projectID,
	)

	if err != nil {
		return err
	}

	if !allowed {
		return domain.ErrForbidden
	}

	return u.Delete(ctx, projectID)
}