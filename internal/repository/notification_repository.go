package repository

import (
	"context"

	"backend-api/internal/domain"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *domain.Notification) error
}
