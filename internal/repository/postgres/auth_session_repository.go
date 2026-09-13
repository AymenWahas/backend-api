package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"backend-api/internal/domain"
)

type AuthSessionRepository struct {
	db *gorm.DB
}

func NewAuthSessionRepository(db *gorm.DB) *AuthSessionRepository {
	return &AuthSessionRepository{
		db: db,
	}
}

func (r *AuthSessionRepository) Create(
	ctx context.Context,
	session domain.AuthSession,
) error {
	return r.db.WithContext(ctx).Create(&session).Error
}

func (r *AuthSessionRepository) GetByRefreshTokenHash(
	ctx context.Context,
	hash string,
) (domain.AuthSession, error) {
	var session domain.AuthSession

	result := r.db.WithContext(ctx).
		Where("refresh_token_hash = ?", hash).
		First(&session)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return domain.AuthSession{}, domain.ErrSessionNotFound
	}

	if result.Error != nil {
		return domain.AuthSession{}, result.Error
	}

	return session, nil
}

func (r *AuthSessionRepository) Revoke(
	ctx context.Context,
	id string,
) error {
	result := r.db.WithContext(ctx).
		Model(&domain.AuthSession{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"revoked_at": gorm.Expr("CURRENT_TIMESTAMP"),
		})

	return result.Error
}
func (r *AuthSessionRepository) Rotate(
	ctx context.Context,
	oldSessionID string,
	newSession domain.AuthSession,
) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&domain.AuthSession{}).
			Where("id = ?", oldSessionID).
			Updates(map[string]interface{}{
				"revoked_at":  gorm.Expr("CURRENT_TIMESTAMP"),
				"replaced_by": newSession.ID,
			})

		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected != 1 {
			return domain.ErrSessionNotFound
		}

		if err := tx.Create(&newSession).Error; err != nil {
			return err
		}

		return nil
	})
}
