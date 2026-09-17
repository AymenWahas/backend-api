package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"backend-api/internal/cache"
	"backend-api/internal/config"
	"backend-api/internal/database"
	httpRouter "backend-api/internal/delivery/http"
	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/messaging"
	"backend-api/internal/observability"
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
	observability.Register()

	// Application context.
	// It is cancelled when the process receives SIGTERM or Ctrl+C.
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	// Configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"configuration validation failed",
			"error",
			err,
		)
		return
	}

	// Database
	db, err := database.NewPostgres(cfg)
	if err != nil {
		slog.Error(
			"database connection failed",
			"error",
			err,
		)
		return
	}

	observability.RecordDBStats(database.Stats(db))
	//to update connection db
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				observability.RecordDBStats(database.Stats(db))

			case <-ctx.Done():
				return
			}
		}
	}()

	// Redis
	redisClient := cache.NewRedis("localhost:6379")

	if err := redisClient.Ping(ctx); err != nil {
		slog.Error(
			"redis connection failed",
			"error",
			err,
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
	membershipRepo := postgres.NewMembershipRepository(db)

	// Authorization
	authorizationUC := usecase.NewAuthorizationUsecase(
		userRepo,
		projectRepo,
		membershipRepo,
		taskRepo,
	)

	// Usecases
	employeeUC := usecase.NewEmployeeUsecase(
		employeeRepo,
	)

	projectUC := usecase.NewProjectUsecase(
		projectRepo,
		redisClient,
		authorizationUC,
	)

	taskUC := usecase.NewTaskUsecase(
		taskRepo,
		eventPublisher,
		authorizationUC,
	)

	authUC := usecase.NewAuthUsecase(
		employeeRepo,
		authSessionRepo,
		userRepo,
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
	)

	projectMembershipUC := usecase.NewProjectMembershipUsecase(
		projectRepo,
		membershipRepo,
		employeeRepo,
		authorizationUC,
	)

	// Notification Worker
	notificationWorker := worker.NewNotificationWorker(
		redisClient.Client(),
		notificationRepo,
		slog.Default(),
	)

	var workerWG sync.WaitGroup
	workerWG.Add(1)

	go func() {
		defer workerWG.Done()

		if err := notificationWorker.Run(ctx); err != nil &&
			!errors.Is(err, context.Canceled) {

			slog.Error(
				"notification worker stopped with error",
				"error",
				err,
			)
		}
	}()

	// HTTP Handlers
	h := handler.NewHandler(
		employeeUC,
		projectUC,
		taskUC,
		projectMembershipUC,
	)

	authHandler := handler.NewAuthHandler(authUC)

	// Router
	router := httpRouter.NewRouter(
		h,
		authHandler,
		cfg.RequestTimeout,
		cfg.JWTSecret,
		cfg.AllowedOrigins,
	)

	// HTTP Server
	port := cfg.Port
	addr := ":" + port

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	slog.Info(
		"HTTPS server starting",
		"addr",
		addr,
	)

	go func() {
		if err := server.ListenAndServeTLS(
			"certs/cert.pem",
			"certs/key.pem",
		); err != nil && err != http.ErrServerClosed {
			slog.Error(
				"server failed",
				"error",
				err,
			)

			stop()
		}
	}()

	// Wait for SIGTERM or Ctrl+C.
	<-ctx.Done()

	slog.Info("shutdown signal received")

	// Give active HTTP requests time to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error(
			"HTTP server shutdown failed",
			"error",
			err,
		)
	} else {
		slog.Info("HTTP server shutdown complete")
	}
	workerWG.Wait()

	// ctx was cancelled by SIGTERM/Ctrl+C.
	// The notification worker receives the same cancellation
	// through ctx and should stop its processing.
	slog.Info("application shutdown complete")
}
