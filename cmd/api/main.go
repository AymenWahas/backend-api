package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"backend-api/internal/cache"
	"backend-api/internal/config"
	"backend-api/internal/database"
	httpRouter "backend-api/internal/delivery/http"
	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/messaging"
	"backend-api/internal/repository/postgres"
	"backend-api/internal/usecase"
	"backend-api/internal/worker"
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

	// Redis
	redisClient := cache.NewRedis("localhost:6379")

	if err := redisClient.Ping(context.Background()); err != nil {
		slog.Error(
			"redis connection failed",
			"error", err,
		)
		return
	}

	slog.Info("redis connection successful")

	// Event Publisher
	eventPublisher := messaging.NewRedisStreamPublisher(
		redisClient.Client(),
	)

	// Repositories
	employeeRepo := postgres.NewEmployeeRepository(db)
	projectRepo := postgres.NewProjectRepository(db)
	taskRepo := postgres.NewTaskRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	userRepo := postgres.NewUserRepository(db)
	authSessionRepo := postgres.NewAuthSessionRepository(db)

	// Usecases
	employeeUC := usecase.NewEmployeeUsecase(
		employeeRepo,
	)

	projectUC := usecase.NewProjectUsecase(
		projectRepo,
		redisClient,
	)

	taskUC := usecase.NewTaskUsecase(
		taskRepo,
		eventPublisher,
	)

	authUC := usecase.NewAuthUsecase(
		employeeRepo,
		authSessionRepo,
		userRepo,
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
	)

	// Notification Worker
	notificationWorker := worker.NewNotificationWorker(
		redisClient.Client(),
		notificationRepo,
		slog.Default(),
	)

	go func() {
		if err := notificationWorker.Run(context.Background()); err != nil {
			slog.Error(
				"notification worker stopped",
				"error", err,
			)
		}
	}()

	// HTTP Handlers
	h := handler.NewHandler(
		employeeUC,
		projectUC,
		taskUC,
	)

	authHandler := handler.NewAuthHandler(authUC)

	// Router
	router := httpRouter.NewRouter(
		h,
		authHandler,
		cfg.RequestTimeout,
		cfg.JWTSecret,
	)

	// Port
	port := cfg.Port
	addr := ":" + port

	slog.Info(
		"HTTPS server starting",
		"addr", addr,
	)

	// HTTPS Server
	err = http.ListenAndServeTLS(
		addr,
		"certs/cert.pem",
		"certs/key.pem",
		router,
	)

	if err != nil {
		slog.Error(
			"server failed",
			"error", err,
		)
	}
}
