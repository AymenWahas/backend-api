package main

import (
	"log/slog"
	"net/http"
	"os"

	"backend-api/internal/config"
	"backend-api/internal/database"
	httpdelivery "backend-api/internal/delivery/http"
	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/repository/postgres"
	"backend-api/internal/usecase"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	slog.SetDefault(logger)

	cfg, err := config.Load()

	if err != nil {
		slog.Error(
			"configuration validation failed",
			"error", err,
		)

		return
	}

	// Database
	db, err := database.NewPostgres(cfg)

	if err != nil {
		slog.Error(
			"database connection failed",
			"error", err,
		)

		return
	}

	// Repository
	repo := postgres.NewEmployeeRepository(db)
	employeeUC := usecase.NewEmployeeUsecase(repo)

	projectRepo := postgres.NewProjectRepository(db)
	projectUC := usecase.NewProjectUsecase(projectRepo)

	taskRepo := postgres.NewTaskRepository(db)
	taskUC := usecase.NewTaskUsecase(taskRepo)

	h := handler.NewHandler(employeeUC, projectUC, taskUC)
	// Router
	router := httpdelivery.NewRouter(h)

	// Port
	port := cfg.Port

	addr := ":" + port

	slog.Info("HTTPS server starting", "addr", addr)

	// HTTPS Server
	err = http.ListenAndServeTLS(
		addr,
		"certs/cert.pem",
		"certs/key.pem",
		router,
	)

	if err != nil {
		slog.Error("server failed", "error", err)
	}
}
