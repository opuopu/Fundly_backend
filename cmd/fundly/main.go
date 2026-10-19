package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dodul01/fundly/internal/config"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
	"github.com/Dodul01/fundly/internal/infrastructure/persistence/postgres"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Initialize PostgreSQL Connection Pool
	dbCfg := postgres.DefaultConfig(cfg.DatabaseURL)
	pool, err := postgres.NewPool(ctx, dbCfg)
	if err != nil {
		log.Printf("⚠️ PostgreSQL connection failed: %v", err)
		log.Printf("💡 To start PostgreSQL: run 'docker compose up -d' or configure DATABASE_URL in .env")
	} else {
		defer pool.Close()
		log.Println("✅ Connected to PostgreSQL database successfully.")

		// 2. Run Database Migrations if configured
		if cfg.AutoMigrate {
			if err := postgres.ApplyMigrations(ctx, pool); err != nil {
				log.Printf("⚠️ Failed to apply database migrations: %v", err)
			} else {
				log.Println("✅ Database migrations applied successfully.")
			}
		}
	}

	// 3. Initialize HTTP Router
	mux := infrahttp.BuildRouter(pool, cfg)

	serverAddr := ":" + cfg.Port
	server := &http.Server{
		Addr:    serverAddr,
		Handler: mux,
	}

	// Graceful shutdown handling
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("Fundly server listening on %s (env: %s)...\n", serverAddr, cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed to start: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down HTTP server: %v", err)
	}
	log.Println("Fundly server stopped.")
}