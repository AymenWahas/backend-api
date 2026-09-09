package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"backend-api/internal/database"
	"backend-api/internal/domain"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{
		db: db,
	}
}

func mapNotificationDBError(err error) error {
	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" {
			return domain.ErrNotificationAlreadyExists
		}
	}

	return err
}

func (r *NotificationRepository) Create(
	ctx context.Context,
	notification *domain.Notification,
) error {
	result := database.DBFromContext(ctx, r.db).
		WithContext(ctx).
		Create(notification)

	if result.Error != nil {
		return mapNotificationDBError(result.Error)
	}

	return nil
}
