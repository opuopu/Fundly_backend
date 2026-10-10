package http

import (
	"encoding/json"
	"net/http"
)

type Handler struct {
	// Application services will be injected here (e.g. transferService *transfer.Service)
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"app":    "fundly",
	})
}
