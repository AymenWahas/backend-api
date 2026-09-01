package main

import (
	"fmt"
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

	fmt.Println("HTTPS server running on", addr)

	err := http.ListenAndServeTLS(
		addr,
		"certs/cert.pem",
		"certs/key.pem",
		nil,
	)

	if err != nil {
		fmt.Println("Server error:", err)
	}
	fmt.Println("Hello Backend")
	fmt.Println("Hello Backend 3")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"status":"ok"}`)
}
