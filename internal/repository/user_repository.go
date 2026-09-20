package repository

import (
	"context"

	"backend-api/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByID(ctx context.Context, id int) (domain.User, error)
	GetByEmployeeID(ctx context.Context, employeeID int) (domain.User, error)
}
