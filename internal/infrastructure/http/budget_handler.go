package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	budgetapp "github.com/Dodul01/fundly/internal/application/budget"
	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
)

type BudgetHandler struct {
	listUC   *budgetapp.ListBudgetsUseCase
	createUC *budgetapp.CreateBudgetUseCase
	getUC    *budgetapp.GetBudgetUseCase
	updateUC *budgetapp.UpdateBudgetUseCase
	deleteUC *budgetapp.DeleteBudgetUseCase
}

func NewBudgetHandler(
	listUC *budgetapp.ListBudgetsUseCase,
	createUC *budgetapp.CreateBudgetUseCase,
	getUC *budgetapp.GetBudgetUseCase,
	updateUC *budgetapp.UpdateBudgetUseCase,
	deleteUC *budgetapp.DeleteBudgetUseCase,
) *BudgetHandler {
	return &BudgetHandler{
		listUC:   listUC,
		createUC: createUC,
		getUC:    getUC,
		updateUC: updateUC,
		deleteUC: deleteUC,
	}
}

// HandleBudgets routes /api/v1/budgets.
func (h *BudgetHandler) HandleBudgets(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListBudgets(w, r)
	case http.MethodPost:
		h.CreateBudget(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// HandleBudgetByID routes /api/v1/budgets/{id}.
func (h *BudgetHandler) HandleBudgetByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetBudget(w, r)
	case http.MethodPatch:
		h.UpdateBudget(w, r)
	case http.MethodDelete:
		h.DeleteBudget(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// ListBudgets handles GET /api/v1/budgets.
func (h *BudgetHandler) ListBudgets(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	q := r.URL.Query()
	items, err := h.listUC.Execute(r.Context(), userID, q.Get("month"), q.Get("category_id"))
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve budgets")
		return
	}

	JSON(w, http.StatusOK, items, "")
}

type CreateBudgetRequest struct {
	Month       string `json:"month"`
	LimitAmount int64  `json:"limit_amount"`
	CategoryID  string `json:"category_id"`
	Currency    string `json:"currency"`
	Note        string `json:"note"`
	Description string `json:"description"`
}

// CreateBudget handles POST /api/v1/budgets.
func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.createUC.Execute(r.Context(), budgetapp.CreateBudgetInput{
		UserID:      userID,
		Month:       req.Month,
		LimitAmount: req.LimitAmount,
		CategoryID:  req.CategoryID,
		Currency:    req.Currency,
		Note:        req.Note,
		Description: req.Description,
	})
	if err != nil {
		writeBudgetError(w, err, "Failed to create budget")
		return
	}

	JSON(w, http.StatusCreated, res, "Budget created successfully")
}

// GetBudget handles GET /api/v1/budgets/{id}.
func (h *BudgetHandler) GetBudget(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractBudgetID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_BUDGET_ID", "Budget ID is required")
		return
	}

	res, err := h.getUC.Execute(r.Context(), userID, id)
	if err != nil {
		writeBudgetError(w, err, "Failed to retrieve budget")
		return
	}

	JSON(w, http.StatusOK, res, "")
}

type UpdateBudgetRequest struct {
	LimitAmount *int64  `json:"limit_amount"`
	Note        *string `json:"note"`
	Description *string `json:"description"`
}

// UpdateBudget handles PATCH /api/v1/budgets/{id}.
func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractBudgetID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_BUDGET_ID", "Budget ID is required")
		return
	}

	var req UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.updateUC.Execute(r.Context(), budgetapp.UpdateBudgetInput{
		UserID:      userID,
		ID:          id,
		LimitAmount: req.LimitAmount,
		Note:        req.Note,
		Description: req.Description,
	})
	if err != nil {
		writeBudgetError(w, err, "Failed to update budget")
		return
	}

	JSON(w, http.StatusOK, res, "Budget updated successfully")
}

// DeleteBudget handles DELETE /api/v1/budgets/{id}.
func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractBudgetID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_BUDGET_ID", "Budget ID is required")
		return
	}

	if err := h.deleteUC.Execute(r.Context(), userID, id); err != nil {
		writeBudgetError(w, err, "Failed to delete budget")
		return
	}

	JSON(w, http.StatusOK, nil, "Budget deleted successfully")
}

func writeBudgetError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, domainBudget.ErrBudgetNotFound):
		JSONError(w, http.StatusNotFound, "BUDGET_NOT_FOUND", "Budget not found")
	case errors.Is(err, domainBudget.ErrBudgetExists):
		JSONError(w, http.StatusConflict, "BUDGET_EXISTS", "A budget already exists for this category and month")
	case errors.Is(err, domainBudget.ErrInvalidMonth):
		JSONError(w, http.StatusBadRequest, "INVALID_MONTH", "Month must be in YYYY-MM format")
	case errors.Is(err, domainBudget.ErrInvalidLimit):
		JSONError(w, http.StatusBadRequest, "INVALID_LIMIT", "Limit amount must be greater than zero")
	case errors.Is(err, domainBudget.ErrCategoryRequired):
		JSONError(w, http.StatusBadRequest, "CATEGORY_REQUIRED", "category_id is required")
	case errors.Is(err, domainBudget.ErrNoFieldsToUpdate):
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "At least one field must be provided to update")
	default:
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fallback)
	}
}

func extractBudgetID(r *http.Request) string {
	id := r.PathValue("id")
	if id != "" {
		return id
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "budgets" {
		return parts[3]
	}
	return ""
}
