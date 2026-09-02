package middleware

import (
	"net/http"
	"strings"
)

func AcceptJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")

		if accept != "" &&
			accept != "*/*" &&
			!strings.Contains(accept, "application/json") {

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotAcceptable)

			w.Write([]byte(`{
				"error": {
					"code": "NOT_ACCEPTABLE",
					"message": "only application/json is supported"
				}
			}`))

			return
		}

		next.ServeHTTP(w, r)
	})
}
