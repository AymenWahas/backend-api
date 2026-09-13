package domain

import "time"

type AuthSession struct {
	ID               string     `json:"id"`
	UserID           int        `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
	ReplacedBy       *string    `json:"replaced_by,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}
