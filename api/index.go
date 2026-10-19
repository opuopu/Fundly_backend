package handler

import (
	"context"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/Dodul01/fundly/internal/config"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
	"github.com/Dodul01/fundly/internal/infrastructure/persistence/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	once   sync.Once
	router http.Handler
	pool   *pgxpool.Pool
)

func initApp() {
	cfg := config.Load()

	// In serverless environments, only /tmp is writable
	if cfg.UploadsDir == "./uploads" || cfg.UploadsDir == "" {
		cfg.UploadsDir = "/tmp/uploads"
	}
	_ = os.MkdirAll(cfg.UploadsDir, 0755)

	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		dbCfg := postgres.DefaultConfig(cfg.DatabaseURL)
		var err error
		pool, err = postgres.NewPool(ctx, dbCfg)
		if err == nil && cfg.AutoMigrate {
			_ = postgres.ApplyMigrations(ctx, pool)
		}
	}

	router = infrahttp.BuildRouter(pool, cfg)
}

// Handler is the Vercel Serverless Function entrypoint.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(initApp)

	// If Vercel rewrote the path to /api/index, restore the client's original requested path
	if r.URL.Path == "/api/index" || r.URL.Path == "/api" {
		if orig := r.Header.Get("x-matched-path"); orig != "" {
			r.URL.Path = orig
		} else if orig := r.Header.Get("x-forwarded-uri"); orig != "" {
			r.URL.Path = orig
		}
	}

	router.ServeHTTP(w, r)
}
