package usecase

import (
	"context"
	"errors"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
	"backend-api/internal/security"
)

type AuthorizationUsecase struct {
	userRepo       repository.UserRepository
	projectRepo    repository.ProjectRepository
	membershipRepo repository.MembershipRepository
	taskRepo       repository.TaskRepository
}

func NewAuthorizationUsecase(
	userRepo repository.UserRepository,
	projectRepo repository.ProjectRepository,
	membershipRepo repository.MembershipRepository,
	taskRepo repository.TaskRepository,
) *AuthorizationUsecase {
	return &AuthorizationUsecase{
		userRepo:       userRepo,
		projectRepo:    projectRepo,
		membershipRepo: membershipRepo,
		taskRepo:       taskRepo,
	}
}

func (u *AuthorizationUsecase) getProjectAccess(
	ctx context.Context,
	userID int,
	projectID uint,
) (
	domain.User,
	domain.Project,
	*domain.Membership,
	bool,
	error,
) {
	user, err := u.userRepo.GetByID(ctx, userID)
	if err != nil {
		return domain.User{}, domain.Project{}, nil, false, err
	}

	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return domain.User{}, domain.Project{}, nil, false, err
	}

	// Project owner does not need a membership row.
	if project.OwnerID == user.EmployeeID {
		return user, *project, nil, true, nil
	}

	membership, err := u.membershipRepo.GetByEmployeeProject(
		ctx,
		user.EmployeeID,
		projectID,
	)
	if err != nil {
		if errors.Is(err, domain.ErrMembershipNotFound) {
			return user, *project, nil, false, nil
		}

		return domain.User{}, domain.Project{}, nil, false, err
	}

	return user, *project, &membership, false, nil
}

func (u *AuthorizationUsecase) CanAccessProject(
	ctx context.Context,
	userID int,
	projectID uint,
) (bool, error) {
	_, _, membership, isOwner, err :=
		u.getProjectAccess(ctx, userID, projectID)

	if err != nil {
		return false, err
	}

	if isOwner {
		return true, nil
	}

	if membership == nil {
		return false, nil
	}

	return security.CanAccessProject(
		security.Role(membership.Role),
		true,
		membership.IsActive,
	), nil
}

func (u *AuthorizationUsecase) CanModifyProject(
	ctx context.Context,
	userID int,
	projectID uint,
) (bool, error) {
	_, _, membership, isOwner, err :=
		u.getProjectAccess(ctx, userID, projectID)

	if err != nil {
		return false, err
	}

	if isOwner {
		return true, nil
	}

	if membership == nil || !membership.IsActive {
		return false, nil
	}

	return security.CanModifyProject(
		security.Role(membership.Role),
		false,
	), nil
}

func (u *AuthorizationUsecase) CanDeleteProject(
	ctx context.Context,
	userID int,
	projectID uint,
) (bool, error) {
	_, _, membership, isOwner, err :=
		u.getProjectAccess(ctx, userID, projectID)

	if err != nil {
		return false, err
	}

	if isOwner {
		return true, nil
	}

	if membership == nil || !membership.IsActive {
		return false, nil
	}

	return security.CanDeleteProject(
		security.Role(membership.Role),
		false,
	), nil
}

func (u *AuthorizationUsecase) CanManageMembers(
	ctx context.Context,
	userID int,
	projectID uint,
) (bool, error) {
	_, _, membership, isOwner, err :=
		u.getProjectAccess(ctx, userID, projectID)

	if err != nil {
		return false, err
	}

	if isOwner {
		return true, nil
	}

	if membership == nil || !membership.IsActive {
		return false, nil
	}

	return security.CanManageMembers(
		security.Role(membership.Role),
		false,
	), nil
}

func (u *AuthorizationUsecase) CanAccessTask(
	ctx context.Context,
	userID int,
	taskID uint,
) (bool, error) {
	if u.taskRepo == nil {
		return false, errors.New("task repository unavailable")
	}

	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return false, err
	}

	return u.CanAccessProject(
		ctx,
		userID,
		task.ProjectID,
	)
}

func (u *AuthorizationUsecase) CanModifyTask(
	ctx context.Context,
	userID int,
	taskID uint,
) (bool, error) {
	if u.taskRepo == nil {
		return false, errors.New("task repository unavailable")
	}

	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return false, err
	}

	_, _, membership, isOwner, err :=
		u.getProjectAccess(ctx, userID, task.ProjectID)

	if err != nil {
		return false, err
	}

	if isOwner {
		return security.CanModifyTask(
			security.RoleOwner,
			true,
		), nil
	}

	if membership == nil || !membership.IsActive {
		return false, nil
	}

	return security.CanModifyTask(
		security.Role(membership.Role),
		false,
	), nil
}

func (u *AuthorizationUsecase) CanDeleteTask(
	ctx context.Context,
	userID int,
	taskID uint,
) (bool, error) {
	if u.taskRepo == nil {
		return false, errors.New("task repository unavailable")
	}

	task, err := u.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return false, err
	}

	_, _, membership, isOwner, err :=
		u.getProjectAccess(ctx, userID, task.ProjectID)

	if err != nil {
		return false, err
	}

	if isOwner {
		return security.CanDeleteTask(
			security.RoleOwner,
			true,
		), nil
	}

	if membership == nil || !membership.IsActive {
		return false, nil
	}

	return security.CanDeleteTask(
		security.Role(membership.Role),
		false,
	), nil
}
