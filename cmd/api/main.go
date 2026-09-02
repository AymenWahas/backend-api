package main

import (
	"log/slog"
	"net/http"
	"time"

	"backend-api/internal/config"
	"backend-api/internal/database"
	httpdelivery "backend-api/internal/delivery/http"
	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/repository/postgres"
	"backend-api/internal/usecase"
)

func main() {
	cfg := config.Load()

	// Database
	db, err := database.NewPostgres(cfg)
	if err != nil {
		slog.Error("database connection failed", "error", err)
		return
	}
	go func() {
		for {
			stats := database.Stats(db)

			slog.Info(
				"database pool",
				"open", stats.OpenConnections,
				"in_use", stats.InUse,
				"idle", stats.Idle,
			)

			time.Sleep(10 * time.Second)
		}
	}()
	// Repository
	repo := postgres.NewEmployeeRepository(db)

	// Usecase
	uc := usecase.NewEmployeeUsecase(repo)

	// HTTP Handler
	h := handler.NewHandler(uc)

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
