package usecase

import (
	"context"
	"strings"

	"backend-api/internal/domain"
	"backend-api/internal/repository"
)

type ProjectMembershipUsecase struct {
	projectRepo   repository.ProjectRepository
	repo          repository.MembershipRepository
	employeeRepo  repository.EmployeeRepository
	authorization *AuthorizationUsecase
}

func NewProjectMembershipUsecase(
	projectRepo repository.ProjectRepository,
	repo repository.MembershipRepository,
	employeeRepo repository.EmployeeRepository,
	authorization *AuthorizationUsecase,
) *ProjectMembershipUsecase {
	return &ProjectMembershipUsecase{
		projectRepo:   projectRepo,
		repo:          repo,
		employeeRepo:  employeeRepo,
		authorization: authorization,
	}
}

func normalizeProjectRole(raw string) (string, error) {
	role := strings.TrimSpace(strings.ToLower(raw))

	switch role {
	case "member", "manager", "admin":
		return role, nil
	default:
		return "", domain.ErrInvalidProjectRole
	}
}

func (u *ProjectMembershipUsecase) GetMembers(
	ctx context.Context,
	userID int,
	projectID uint,
) ([]domain.Membership, error) {
	allowed, err := u.authorization.CanAccessProject(
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

	return u.repo.GetByProject(ctx, projectID)
}

func (u *ProjectMembershipUsecase) AddMember(
	ctx context.Context,
	userID int,
	projectID uint,
	employeeID int,
	role string,
) error {
	allowed, err := u.authorization.CanManageMembers(
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

	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// The project owner is not a membership role.
	if project.OwnerID == employeeID {
		return domain.ErrInvalidProjectRole
	}

	if _, err := u.employeeRepo.GetByID(ctx, employeeID); err != nil {
		return err
	}

	normalized, err := normalizeProjectRole(role)
	if err != nil {
		return err
	}

	membership := domain.Membership{
		EmployeeID: employeeID,
		ProjectID:  projectID,
		Role:       normalized,
	}

	return u.repo.Create(ctx, membership)
}

func (u *ProjectMembershipUsecase) UpdateRole(
	ctx context.Context,
	userID int,
	projectID uint,
	employeeID int,
	role string,
) error {
	allowed, err := u.authorization.CanManageMembers(
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

	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// The owner does not have a membership role.
	if project.OwnerID == employeeID {
		return domain.ErrInvalidProjectRole
	}

	normalized, err := normalizeProjectRole(role)
	if err != nil {
		return err
	}

	return u.repo.UpdateRole(
		ctx,
		employeeID,
		projectID,
		normalized,
	)
}

func (u *ProjectMembershipUsecase) RemoveMember(
	ctx context.Context,
	userID int,
	projectID uint,
	employeeID int,
) error {
	allowed, err := u.authorization.CanManageMembers(
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

	project, err := u.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		return err
	}

	// The owner cannot be removed as a membership.
	if project.OwnerID == employeeID {
		return domain.ErrForbidden
	}

	return u.repo.Delete(
		ctx,
		employeeID,
		projectID,
	)
}
