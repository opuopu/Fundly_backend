package http

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	pool *pgxpool.Pool
}

func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

type HealthResponse struct {
	Status     string            `json:"status"`
	App        string            `json:"app"`
	Timestamp  string            `json:"timestamp"`
	Components map[string]string `json:"components"`
}

// HealthCheck verifies application and database connectivity.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dbStatus := "down"
	isHealthy := true

	if h.pool != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.pool.Ping(ctx); err == nil {
			dbStatus = "up"
		} else {
			isHealthy = false
		}
	} else {
		dbStatus = "disconnected"
	}

	statusCode := http.StatusOK
	statusText := "healthy"
	if !isHealthy {
		statusCode = http.StatusServiceUnavailable
		statusText = "degraded"
	}

	resp := HealthResponse{
		Status:    statusText,
		App:       "fundly",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Components: map[string]string{
			"database": dbStatus,
		},
	}

	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(resp)
}
