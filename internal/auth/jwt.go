package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessTokenClaims struct {
	UserID     int `json:"user_id"`
	EmployeeID int `json:"employee_id"`

	jwt.RegisteredClaims
}

func GenerateAccessToken(
	userID int,
	employeeID int,
	secret string,
	ttl time.Duration,
) (string, error) {

	now := time.Now()

	claims := AccessTokenClaims{
		UserID:     userID,
		EmployeeID: employeeID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString([]byte(secret))
}