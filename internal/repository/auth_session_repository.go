package repository

import (
	"context"

	"backend-api/internal/domain"
)

type AuthSessionRepository interface {
	Create(
		ctx context.Context,
		session domain.AuthSession,
	) error

	GetByRefreshTokenHash(
		ctx context.Context,
		hash string,
	) (domain.AuthSession, error)

	Revoke(
		ctx context.Context,
		id string,
	) error

	Rotate(
		ctx context.Context,
		oldSessionID string,
		newSession domain.AuthSession,
	) error
}
