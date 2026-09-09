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
}

func NewProjectUsecase(repo repository.ProjectRepository, cache *cache.RedisClient) *ProjectUsecase {
	return &ProjectUsecase{
		repo:  repo,
		cache: cache,
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

	err := u.cache.Get(ctx, key, &cachedProject)
	if err == nil {
		slog.Info(
			"project cache hit",
			"project_id", id,
			"cache_key", key,
		)

		return &cachedProject, nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		slog.Warn(
			"project cache unavailable",
			"project_id", id,
			"cache_key", key,
			"error", err,
		)
	} else {
		slog.Info(
			"project cache miss",
			"project_id", id,
			"cache_key", key,
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
			"project_id", id,
			"cache_key", key,
			"error", err,
		)

		return project, nil
	}

	slog.Info(
		"project cached",
		"project_id", id,
		"cache_key", key,
		"ttl", projectCacheTTL.String(),
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
			"project_id", project.ID,
			"version", project.Version,
			"error", err,
		)

		return err
	}

	key := fmt.Sprintf("project:%d", project.ID)

	if err := u.cache.Delete(ctx, key); err != nil {
		slog.Warn(
			"project cache invalidation failed",
			"project_id", project.ID,
			"cache_key", key,
			"error", err,
		)

		return nil
	}

	slog.Info(
		"project cache invalidated",
		"project_id", project.ID,
		"cache_key", key,
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

	key := fmt.Sprintf("project:%d", id)

	if err := u.cache.Delete(ctx, key); err != nil {
		slog.Warn(
			"project cache invalidation failed",
			"project_id", id,
			"cache_key", key,
			"error", err,
		)

		return nil
	}

	slog.Info(
		"project cache invalidated",
		"project_id", id,
		"cache_key", key,
	)

	return nil
}
