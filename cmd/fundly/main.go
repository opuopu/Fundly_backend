package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Dodul01/fundly/internal/config"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

func main() {
	cfg := config.Load()

	handler := infrahttp.NewHandler()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", handler.HealthCheck)

	serverAddr := ":" + cfg.Port
	fmt.Printf("Fundly server listening on %s (env: %s)...\n", serverAddr, cfg.Environment)

	if err := http.ListenAndServe(serverAddr, mux); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server failed to start: %v", err)
	}
}