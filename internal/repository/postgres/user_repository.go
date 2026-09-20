package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"backend-api/internal/domain"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(
	ctx context.Context,
	user domain.User,
) (domain.User, error) {
	result := r.db.WithContext(ctx).Create(&user)

	if result.Error != nil {
		var pgErr *pgconn.PgError

		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, domain.ErrUserAlreadyExists
		}

		return domain.User{}, result.Error
	}

	return user, nil
}

func (r *UserRepository) GetByEmployeeID(
	ctx context.Context,
	employeeID int,
) (domain.User, error) {
	var user domain.User

	result := r.db.WithContext(ctx).
		Where("employee_id = ?", employeeID).
		First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}

	if result.Error != nil {
		return domain.User{}, result.Error
	}

	return user, nil
}
func (r *UserRepository) GetByID(
	ctx context.Context,
	id int,
) (domain.User, error) {

	var user domain.User

	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&user)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.User{}, domain.ErrUserNotFound
	}

	if result.Error != nil {
		return domain.User{}, result.Error
	}

	return user, nil
}
