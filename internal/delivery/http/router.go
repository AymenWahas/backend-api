package http

import (
	"net/http"

	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/delivery/http/middleware"
)

func NewRouter(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)

	mux.Handle(
		"POST /api/v1/employees",
		middleware.JSONContentType(
			http.HandlerFunc(h.CreateEmployee),
		),
	)

	mux.HandleFunc("GET /api/v1/employees", h.GetEmployees)

	mux.HandleFunc("GET /api/v1/employees/{id}", h.GetEmployee)

	mux.Handle(
		"PUT /api/v1/employees/{id}",
		middleware.JSONContentType(
			http.HandlerFunc(h.UpdateEmployee),
		),
	)

	mux.HandleFunc("DELETE /api/v1/employees/{id}", h.DeleteEmployee)

	return middleware.AcceptJSON(mux)
}
