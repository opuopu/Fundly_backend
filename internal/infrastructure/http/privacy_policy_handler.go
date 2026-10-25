package http

import (
	"encoding/json"
	"errors"
	"net/http"

	privacypolicyapp "github.com/Dodul01/fundly/internal/application/privacypolicy"
	domainPrivacyPolicy "github.com/Dodul01/fundly/internal/domain/privacypolicy"
)

type PrivacyPolicyHandler struct {
	getLatestUC   *privacypolicyapp.GetLatestUseCase
	acknowledgeUC *privacypolicyapp.AcknowledgeUseCase
}

func NewPrivacyPolicyHandler(getLatestUC *privacypolicyapp.GetLatestUseCase, acknowledgeUC *privacypolicyapp.AcknowledgeUseCase) *PrivacyPolicyHandler {
	return &PrivacyPolicyHandler{getLatestUC: getLatestUC, acknowledgeUC: acknowledgeUC}
}

// GetLatest handles GET /api/v1/privacy-policy/latest (public).
func (h *PrivacyPolicyHandler) GetLatest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	res, err := h.getLatestUC.Execute(r.Context())
	if err != nil {
		if errors.Is(err, domainPrivacyPolicy.ErrPolicyNotFound) {
			JSONError(w, http.StatusNotFound, "PRIVACY_POLICY_NOT_FOUND", "No privacy policy is published")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve privacy policy")
		return
	}

	JSON(w, http.StatusOK, res, "")
}

type AcknowledgePrivacyPolicyRequest struct {
	Version string `json:"version"`
}

// Acknowledge handles POST /api/v1/privacy-policy/acknowledge (authenticated).
func (h *PrivacyPolicyHandler) Acknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req AcknowledgePrivacyPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	err := h.acknowledgeUC.Execute(r.Context(), privacypolicyapp.AcknowledgeInput{
		UserID:  userID,
		Version: req.Version,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainPrivacyPolicy.ErrVersionMissing):
			JSONError(w, http.StatusBadRequest, "VERSION_REQUIRED", "version is required")
		case errors.Is(err, domainPrivacyPolicy.ErrPolicyNotFound):
			JSONError(w, http.StatusNotFound, "PRIVACY_POLICY_NOT_FOUND", "The specified privacy policy version does not exist")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to acknowledge privacy policy")
		}
		return
	}

	JSON(w, http.StatusOK, nil, "Privacy policy acknowledged")
}
