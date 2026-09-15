package http

import (
	"net/http"
	"time"

	swaggerui "github.com/swaggest/swgui/v5"

	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/delivery/http/middleware"
)

func NewRouter(
	h *handler.Handler,
	authHandler *handler.AuthHandler,
	requestTimeout time.Duration,
	jwtSecret string,
	allowedOrigins string,
) http.Handler {

	// API routes
	apiMux := http.NewServeMux()

	apiMux.HandleFunc("GET /health", h.Health)

	// Auth
	apiMux.Handle(
		"POST /api/v1/auth/register",
		middleware.JSONContentType(
			http.HandlerFunc(authHandler.Register),
		),
	)

	apiMux.Handle(
		"POST /api/v1/auth/login",
		middleware.JSONContentType(
			http.HandlerFunc(authHandler.Login),
		),
	)

	apiMux.Handle(
		"POST /api/v1/auth/refresh",
		middleware.JSONContentType(
			http.HandlerFunc(authHandler.Refresh),
		),
	)

	apiMux.Handle(
		"POST /api/v1/auth/logout",
		middleware.JSONContentType(
			http.HandlerFunc(authHandler.Logout),
		),
	)

	apiMux.Handle(
		"GET /api/v1/me",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(authHandler.Me),
		),
	)

	// Employee
	apiMux.Handle(
		"POST /api/v1/employees",
		middleware.JSONContentType(
			http.HandlerFunc(h.CreateEmployee),
		),
	)

	apiMux.HandleFunc(
		"GET /api/v1/employees",
		h.GetEmployees,
	)

	apiMux.HandleFunc(
		"GET /api/v1/employees/{id}",
		h.GetEmployee,
	)

	apiMux.Handle(
		"PUT /api/v1/employees/{id}",
		middleware.JSONContentType(
			http.HandlerFunc(h.UpdateEmployee),
		),
	)

	apiMux.HandleFunc(
		"DELETE /api/v1/employees/{id}",
		h.DeleteEmployee,
	)

	// Projects
	apiMux.Handle(
		"POST /api/v1/projects",
		middleware.Auth(jwtSecret)(
			middleware.JSONContentType(
				http.HandlerFunc(h.CreateProject),
			),
		),
	)

	apiMux.Handle(
		"GET /api/v1/projects",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(h.GetProjects),
		),
	)

	apiMux.Handle(
		"GET /api/v1/projects/{id}",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(h.GetProject),
		),
	)

	apiMux.Handle(
		"GET /api/v1/projects/{id}/members",
		middleware.Auth(jwtSecret)(
			middleware.JSONContentType(
				http.HandlerFunc(h.GetProjectMembers),
			),
		),
	)

	apiMux.Handle(
		"POST /api/v1/projects/{id}/members",
		middleware.Auth(jwtSecret)(
			middleware.JSONContentType(
				http.HandlerFunc(h.AddProjectMember),
			),
		),
	)

	apiMux.Handle(
		"PATCH /api/v1/projects/{id}/members/{employeeId}/role",
		middleware.Auth(jwtSecret)(
			middleware.JSONContentType(
				http.HandlerFunc(h.UpdateProjectMemberRole),
			),
		),
	)

	apiMux.Handle(
		"DELETE /api/v1/projects/{id}/members/{employeeId}",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(h.RemoveProjectMember),
		),
	)

	apiMux.Handle(
		"PUT /api/v1/projects/{id}",
		middleware.Auth(jwtSecret)(
			middleware.JSONContentType(
				http.HandlerFunc(h.UpdateProject),
			),
		),
	)

	apiMux.Handle(
		"DELETE /api/v1/projects/{id}",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(h.DeleteProject),
		),
	)

	// Tasks
	apiMux.Handle(
		"POST /api/v1/tasks",
		middleware.Auth(jwtSecret)(
			middleware.JSONContentType(
				http.HandlerFunc(h.CreateTask),
			),
		),
	)

	apiMux.Handle(
		"GET /api/v1/tasks",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(h.GetTasks),
		),
	)

	apiMux.Handle(
		"GET /api/v1/tasks/{id}",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(h.GetTask),
		),
	)

	apiMux.Handle(
		"PUT /api/v1/tasks/{id}",
		middleware.Auth(jwtSecret)(
			middleware.JSONContentType(
				http.HandlerFunc(h.UpdateTask),
			),
		),
	)

	apiMux.Handle(
		"DELETE /api/v1/tasks/{id}",
		middleware.Auth(jwtSecret)(
			http.HandlerFunc(h.DeleteTask),
		),
	)

	// Main router
	mux := http.NewServeMux()

	// API
	mux.Handle(
		"/",
		middleware.AcceptJSON(apiMux),
	)

	// OpenAPI specification
	mux.HandleFunc(
		"GET /openapi.json",
		func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(
				w,
				r,
				"docs/openapi.json",
			)
		},
	)

	// Swagger UI
	swaggerHandler := swaggerui.New(
		"Employee API",
		"/openapi.json",
		"/swagger/",
	)

	mux.Handle(
		"/swagger/",
		swaggerHandler,
	)

	// Global middleware
	return middleware.SecurityHeaders(
		middleware.CORS(allowedOrigins)(
			middleware.RateLimit(
				20,
				time.Minute,
			)(
				middleware.RequestSizeLimit(
					1 << 20,
				)(
					middleware.RequestID(
						middleware.Recovery(
							middleware.RequestLogger(
								middleware.Timeout(
									requestTimeout,
								)(mux),
							),
						),
					),
				),
			),
		),
	)
}
