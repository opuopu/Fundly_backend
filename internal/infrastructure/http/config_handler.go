package http

import (
	"encoding/json"
	"errors"
	"net/http"

	configapp "github.com/Dodul01/fundly/internal/application/config"
	domainConfig "github.com/Dodul01/fundly/internal/domain/config"
)

type ConfigHandler struct {
	getUC    *configapp.GetConfigUseCase
	updateUC *configapp.UpdateConfigUseCase
}

func NewConfigHandler(getUC *configapp.GetConfigUseCase, updateUC *configapp.UpdateConfigUseCase) *ConfigHandler {
	return &ConfigHandler{getUC: getUC, updateUC: updateUC}
}

// HandleConfiguration routes /api/v1/configuration.
func (h *ConfigHandler) HandleConfiguration(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.Get(w, r)
	case http.MethodPut:
		h.Update(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// Get handles GET /api/v1/configuration.
func (h *ConfigHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	res, err := h.getUC.Execute(r.Context(), userID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve configuration")
		return
	}

	JSON(w, http.StatusOK, res, "")
}

type UpdateConfigRequest struct {
	DefaultCurrency *string `json:"default_currency"`
	Language        *string `json:"language"`
	Theme           *string `json:"theme"`
}

// Update handles PUT /api/v1/configuration.
func (h *ConfigHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.updateUC.Execute(r.Context(), configapp.UpdateConfigInput{
		UserID:          userID,
		DefaultCurrency: req.DefaultCurrency,
		Language:        req.Language,
		Theme:           req.Theme,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainConfig.ErrInvalidCurrency):
			JSONError(w, http.StatusBadRequest, "INVALID_CURRENCY", "default_currency must be a 3-letter ISO 4217 code")
		case errors.Is(err, domainConfig.ErrNoFieldsToUpdate):
			JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "At least one field must be provided to update")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update configuration")
		}
		return
	}

	JSON(w, http.StatusOK, res, "Configuration updated successfully")
}
