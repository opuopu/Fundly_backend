package http

import (
	"encoding/json"
	"errors"
	"net/http"

	feedbackapp "github.com/Dodul01/fundly/internal/application/feedback"
	domainFeedback "github.com/Dodul01/fundly/internal/domain/feedback"
)

type FeedbackHandler struct {
	submitUC *feedbackapp.SubmitFeedbackUseCase
}

func NewFeedbackHandler(submitUC *feedbackapp.SubmitFeedbackUseCase) *FeedbackHandler {
	return &FeedbackHandler{submitUC: submitUC}
}

type SubmitFeedbackRequest struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
}

// SubmitFeedback handles POST /api/v1/feedback.
func (h *FeedbackHandler) SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req SubmitFeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.submitUC.Execute(r.Context(), feedbackapp.SubmitFeedbackInput{
		UserID:      userID,
		Subject:     req.Subject,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainFeedback.ErrInvalidSubject):
			JSONError(w, http.StatusBadRequest, "INVALID_SUBJECT", "Subject cannot be empty")
		case errors.Is(err, domainFeedback.ErrInvalidDescription):
			JSONError(w, http.StatusBadRequest, "INVALID_DESCRIPTION", "Description cannot be empty")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to submit feedback")
		}
		return
	}

	JSON(w, http.StatusCreated, res, "Thank you for your feedback! Our team has received your submission.")
}
