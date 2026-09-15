package repository

import (
	"backend-api/internal/domain"
	"context"
)

type MembershipRepository interface {
	Create(ctx context.Context, membership domain.Membership) error
	GetByEmployeeProject(ctx context.Context, employeeID int, projectID uint) (domain.Membership, error)
	GetByProject(ctx context.Context, projectID uint) ([]domain.Membership, error)
	UpdateRole(ctx context.Context, employeeID int, projectID uint, role string) error
	Delete(ctx context.Context, employeeID int, projectID uint) error
}
