package middleware

import (
	"net/http"
	"strings"

	"backend-api/internal/delivery/http/response"
)

func JSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")

		if contentType == "" ||
			!strings.HasPrefix(contentType, "application/json") {

			response.WriteError(
				w,
				http.StatusUnsupportedMediaType,
				"UNSUPPORTED_MEDIA_TYPE",
				"Content-Type must be application/json",
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}
