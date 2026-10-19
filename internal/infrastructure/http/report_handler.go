package http

import (
	"encoding/json"
	"errors"
	"net/http"

	reportapp "github.com/Dodul01/fundly/internal/application/report"
)

type ReportHandler struct {
	compareUC *reportapp.CompareUseCase
}

func NewReportHandler(compareUC *reportapp.CompareUseCase) *ReportHandler {
	return &ReportHandler{compareUC: compareUC}
}

type periodRequest struct {
	Month string `json:"month"`
}

type CompareRequest struct {
	Type             string        `json:"type"`
	PrimaryPeriod    periodRequest `json:"primary_period"`
	ComparisonPeriod periodRequest `json:"comparison_period"`
	CategoryIDs      []string      `json:"category_ids"`
}

// Compare handles POST /api/v1/reports/compare.
func (h *ReportHandler) Compare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req CompareRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.compareUC.Execute(r.Context(), reportapp.CompareInput{
		UserID:          userID,
		Type:            req.Type,
		PrimaryMonth:    req.PrimaryPeriod.Month,
		ComparisonMonth: req.ComparisonPeriod.Month,
		CategoryIDs:     req.CategoryIDs,
	})
	if err != nil {
		switch {
		case errors.Is(err, reportapp.ErrInvalidType):
			JSONError(w, http.StatusBadRequest, "INVALID_TYPE", "type must be INCOME or EXPENSE")
		case errors.Is(err, reportapp.ErrInvalidMonth):
			JSONError(w, http.StatusBadRequest, "INVALID_MONTH", "month must be in YYYY-MM format")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to compute comparison")
		}
		return
	}

	JSON(w, http.StatusOK, res, "")
}
