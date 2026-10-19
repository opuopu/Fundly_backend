package http

import (
	"encoding/json"
	"errors"
	"net/http"

	termsapp "github.com/Dodul01/fundly/internal/application/terms"
	domainTerms "github.com/Dodul01/fundly/internal/domain/terms"
)

type TermsHandler struct {
	getLatestUC   *termsapp.GetLatestUseCase
	acknowledgeUC *termsapp.AcknowledgeUseCase
}

func NewTermsHandler(getLatestUC *termsapp.GetLatestUseCase, acknowledgeUC *termsapp.AcknowledgeUseCase) *TermsHandler {
	return &TermsHandler{getLatestUC: getLatestUC, acknowledgeUC: acknowledgeUC}
}

// GetLatest handles GET /api/v1/terms/latest (public).
func (h *TermsHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	res, err := h.getLatestUC.Execute(r.Context())
	if err != nil {
		if errors.Is(err, domainTerms.ErrTermsNotFound) {
			JSONError(w, http.StatusNotFound, "TERMS_NOT_FOUND", "No terms and conditions are published")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve terms")
		return
	}

	JSON(w, http.StatusOK, res, "")
}

type AcknowledgeTermsRequest struct {
	Version string `json:"version"`
}

// Acknowledge handles POST /api/v1/terms/acknowledge (authenticated).
func (h *TermsHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req AcknowledgeTermsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	err := h.acknowledgeUC.Execute(r.Context(), termsapp.AcknowledgeInput{
		UserID:  userID,
		Version: req.Version,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainTerms.ErrVersionMissing):
			JSONError(w, http.StatusBadRequest, "VERSION_REQUIRED", "version is required")
		case errors.Is(err, domainTerms.ErrTermsNotFound):
			JSONError(w, http.StatusNotFound, "TERMS_NOT_FOUND", "The specified terms version does not exist")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to acknowledge terms")
		}
		return
	}

	JSON(w, http.StatusOK, nil, "Terms and conditions acknowledged")
}
