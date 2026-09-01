package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/health", healthHandler)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8443"
	}

	addr := ":" + port

	slog.Info("HTTPS server starting", "addr", addr)

	err := http.ListenAndServeTLS(
		addr,
		"certs/cert.pem",
		"certs/key.pem",
		nil,
	)

	if err != nil {

		slog.Error("server failed", "error", err)
	}

}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info(
		"request received",
		"method", r.Method,
		"path", r.URL.Path,
	)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintln(w, `{"status":"ok"}`)

}
