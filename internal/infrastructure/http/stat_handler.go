package http

import (
	"errors"
	"net/http"

	statapp "github.com/Dodul01/fundly/internal/application/stat"
)

type StatHandler struct {
	breakdownUC *statapp.BreakdownUseCase
	overviewUC  *statapp.OverviewUseCase
}

func NewStatHandler(breakdownUC *statapp.BreakdownUseCase, overviewUC *statapp.OverviewUseCase) *StatHandler {
	return &StatHandler{breakdownUC: breakdownUC, overviewUC: overviewUC}
}

// Breakdown handles GET /api/v1/stats/breakdown.
func (h *StatHandler) Breakdown(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	q := r.URL.Query()
	res, err := h.breakdownUC.Execute(r.Context(), statapp.BreakdownInput{
		UserID:    userID,
		Type:      q.Get("type"),
		Month:     q.Get("month"),
		AccountID: q.Get("account_id"),
	})
	if err != nil {
		if errors.Is(err, statapp.ErrInvalidType) {
			JSONError(w, http.StatusBadRequest, "INVALID_TYPE", "type must be INCOME or EXPENSE")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute breakdown")
		return
	}

	JSON(w, http.StatusOK, res, "")
}

// Overview handles GET /api/v1/stats/overview.
func (h *StatHandler) Overview(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	res, err := h.overviewUC.Execute(r.Context(), userID, r.URL.Query().Get("month"))
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute overview")
		return
	}

	JSON(w, http.StatusOK, res, "")
}
