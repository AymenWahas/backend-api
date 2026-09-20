package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"backend-api/internal/database"
	"backend-api/internal/domain"
)

type MembershipRepository struct {
	db *gorm.DB
}

func NewMembershipRepository(db *gorm.DB) *MembershipRepository {
	return &MembershipRepository{
		db: db,
	}
}

func (r *MembershipRepository) Create(
	ctx context.Context,
	membership domain.Membership,
) error {
	return database.DBFromContext(ctx, r.db).
		WithContext(ctx).
		Create(&membership).
		Error
}

func (r *MembershipRepository) GetByEmployeeProject(
	ctx context.Context,
	employeeID int,
	projectID uint,
) (domain.Membership, error) {

	var membership domain.Membership

	result := database.DBFromContext(ctx, r.db).
		WithContext(ctx).
		Where(
			"employee_id = ? AND project_id = ?",
			employeeID,
			projectID,
		).
		First(&membership)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.Membership{}, domain.ErrMembershipNotFound
	}

	if result.Error != nil {
		return domain.Membership{}, result.Error
	}

	return membership, nil
}

func (r *MembershipRepository) GetByProject(
	ctx context.Context,
	projectID uint,
) ([]domain.Membership, error) {

	var memberships []domain.Membership

	result := database.DBFromContext(ctx, r.db).
		WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&memberships)

	if result.Error != nil {
		return nil, result.Error
	}

	return memberships, nil
}

func (r *MembershipRepository) UpdateRole(
	ctx context.Context,
	employeeID int,
	projectID uint,
	role string,
) error {

	result := database.DBFromContext(ctx, r.db).
		WithContext(ctx).
		Model(&domain.Membership{}).
		Where(
			"employee_id = ? AND project_id = ?",
			employeeID,
			projectID,
		).
		Update("role", role)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrMembershipNotFound
	}

	return nil
}

func (r *MembershipRepository) Delete(
	ctx context.Context,
	employeeID int,
	projectID uint,
) error {

	result := database.DBFromContext(ctx, r.db).
		WithContext(ctx).
		Where(
			"employee_id = ? AND project_id = ?",
			employeeID,
			projectID,
		).
		Delete(&domain.Membership{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrMembershipNotFound
	}

	return nil
}
