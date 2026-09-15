package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"backend-api/internal/auth"
	"backend-api/internal/delivery/http/response"
)

const (
	UserIDContextKey     contextKey = "user_id"
	EmployeeIDContextKey contextKey = "employee_id"
)

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			header := strings.TrimSpace(r.Header.Get("Authorization"))

			if header == "" {
				response.WriteError(
					w,
					http.StatusUnauthorized,
					"UNAUTHORIZED",
					"authorization required",
				)
				return
			}

			parts := strings.Fields(header)

			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.WriteError(
					w,
					http.StatusUnauthorized,
					"INVALID_AUTHORIZATION",
					"invalid authorization header",
				)
				return
			}

			tokenString := parts[1]

			token, err := jwt.ParseWithClaims(
				tokenString,
				&auth.AccessTokenClaims{},
				func(token *jwt.Token) (interface{}, error) {

					if token.Method != jwt.SigningMethodHS256 {
						return nil, errors.New("unexpected signing method")
					}

					return []byte(jwtSecret), nil
				},
			)

			if err != nil || !token.Valid {
				response.WriteError(
					w,
					http.StatusUnauthorized,
					"INVALID_TOKEN",
					"invalid or expired token",
				)
				return
			}

			claims, ok := token.Claims.(*auth.AccessTokenClaims)

			if !ok {
				response.WriteError(
					w,
					http.StatusUnauthorized,
					"INVALID_TOKEN",
					"invalid token claims",
				)
				return
			}

			ctx := context.WithValue(
				r.Context(),
				UserIDContextKey,
				claims.UserID,
			)

			ctx = context.WithValue(
				ctx,
				EmployeeIDContextKey,
				claims.EmployeeID,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		})
	}
}
