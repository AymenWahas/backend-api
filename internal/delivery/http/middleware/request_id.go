package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")

		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(
			r.Context(),
			RequestIDKey,
			requestID,
		)

		w.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(
			w,
			r.WithContext(ctx),
		)
	})
}

func GetRequestID(ctx context.Context) string {
	requestID, _ := ctx.Value(RequestIDKey).(string)
	return requestID
}
