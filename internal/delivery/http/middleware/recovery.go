package middleware

import (
	"log/slog"
	"net/http"

	"backend-api/internal/delivery/http/response"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error(
					"panic recovered",
					"error", err,
					"method", r.Method,
					"path", r.URL.Path,
					"request_id", GetRequestID(r.Context()),
				)

				response.WriteError(
					w,
					http.StatusInternalServerError,
					"INTERNAL_ERROR",
					"internal server error",
				)
			}
		}()

		next.ServeHTTP(w, r)
	})
}
