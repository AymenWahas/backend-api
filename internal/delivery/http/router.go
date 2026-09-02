package http

import (
	"net/http"

	swaggerui "github.com/swaggest/swgui/v5"

	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/delivery/http/middleware"
)

func NewRouter(h *handler.Handler) http.Handler {
	// API routes
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /health", h.Health)

	apiMux.Handle(
		"POST /api/v1/employees",
		middleware.JSONContentType(
			http.HandlerFunc(h.CreateEmployee),
		),
	)

	apiMux.HandleFunc("GET /api/v1/employees", h.GetEmployees)

	apiMux.HandleFunc("GET /api/v1/employees/{id}", h.GetEmployee)

	apiMux.Handle(
		"PUT /api/v1/employees/{id}",
		middleware.JSONContentType(
			http.HandlerFunc(h.UpdateEmployee),
		),
	)

	apiMux.HandleFunc("DELETE /api/v1/employees/{id}", h.DeleteEmployee)

	// Main router
	mux := http.NewServeMux()

	// API goes through AcceptJSON
	mux.Handle("/", middleware.AcceptJSON(apiMux))

	// OpenAPI specification
	mux.HandleFunc("GET /openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.json")
	})

	// Swagger UI
	swaggerHandler := swaggerui.New(
		"Employee API",
		"/openapi.json",
		"/swagger/",
	)

	mux.Handle("/swagger/", swaggerHandler)

	return mux
}
