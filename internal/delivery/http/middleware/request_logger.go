package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"backend-api/internal/observability"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(status int) {
	sr.status = status
	sr.ResponseWriter.WriteHeader(status)
}

func (sr *statusRecorder) Write(body []byte) (int, error) {
	if sr.status == 0 {
		sr.status = http.StatusOK
	}

	return sr.ResponseWriter.Write(body)
}

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		recorder := &statusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)

		duration := time.Since(start)

		observability.RecordHTTP(
			r.Method,
			r.URL.Path,
			recorder.status,
			duration,
		)

		slog.Info(
			"request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration", duration,
			"request_id", GetRequestID(r.Context()),
		)
	})
}
