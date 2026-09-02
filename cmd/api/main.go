package main

import (
	"log/slog"
	"net/http"
	"os"

	httpdelivery "backend-api/internal/delivery/http"
	"backend-api/internal/delivery/http/handler"
	"backend-api/internal/repository/memory"
	"backend-api/internal/usecase"
)

func main() {
	// Repository
	repo := memory.NewEmployeeRepository()

	// Usecase
	uc := usecase.NewEmployeeUsecase(repo)

	// HTTP Handler
	h := handler.NewHandler(uc)

	// Router
	router := httpdelivery.NewRouter(h)

	// Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8443"
	}

	addr := ":" + port

	slog.Info("HTTPS server starting", "addr", addr)

	// HTTPS Server
	err := http.ListenAndServeTLS(
		addr,
		"certs/cert.pem",
		"certs/key.pem",
		router,
	)

	if err != nil {
		slog.Error("server failed", "error", err)
	}
}
