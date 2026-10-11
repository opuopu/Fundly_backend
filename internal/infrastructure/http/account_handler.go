package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	accountapp "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

type AccountHandler struct {
	listAccountsUC  *accountapp.ListAccountsUseCase
	createAccountUC *accountapp.CreateAccountUseCase
	getAccountUC    *accountapp.GetAccountUseCase
	updateAccountUC *accountapp.UpdateAccountUseCase
	deleteAccountUC *accountapp.DeleteAccountUseCase
	getSummaryUC    *accountapp.GetAccountsSummaryUseCase
}

func NewAccountHandler(
	listAccountsUC *accountapp.ListAccountsUseCase,
	createAccountUC *accountapp.CreateAccountUseCase,
	getAccountUC *accountapp.GetAccountUseCase,
	updateAccountUC *accountapp.UpdateAccountUseCase,
	deleteAccountUC *accountapp.DeleteAccountUseCase,
	getSummaryUC *accountapp.GetAccountsSummaryUseCase,
) *AccountHandler {
	return &AccountHandler{
		listAccountsUC:  listAccountsUC,
		createAccountUC: createAccountUC,
		getAccountUC:    getAccountUC,
		updateAccountUC: updateAccountUC,
		deleteAccountUC: deleteAccountUC,
		getSummaryUC:    getSummaryUC,
	}
}

// HandleAccounts routes requests for /api/v1/accounts based on HTTP method.
func (h *AccountHandler) HandleAccounts(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListAccounts(w, r)
	case http.MethodPost:
		h.CreateAccount(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// HandleAccountByID routes requests for /api/v1/accounts/{id} based on HTTP method.
func (h *AccountHandler) HandleAccountByID(w http.ResponseWriter, r *http.Request) {
	if extractAccountID(r) == "summary" {
		h.GetSummary(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.GetAccount(w, r)
	case http.MethodPatch:
		h.UpdateAccount(w, r)
	case http.MethodDelete:
		h.DeleteAccount(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// ListAccounts handles GET /api/v1/accounts.
func (h *AccountHandler) ListAccounts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	groupFilter := r.URL.Query().Get("group")
	accounts, err := h.listAccountsUC.Execute(r.Context(), userID, groupFilter)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve accounts")
		return
	}

	JSON(w, http.StatusOK, accounts, "")
}

type CreateAccountRequest struct {
	Group          string `json:"group"`
	Name           string `json:"name"`
	InitialBalance int64  `json:"initial_balance"`
	Currency       string `json:"currency"`
	Description    string `json:"description"`
}

// CreateAccount handles POST /api/v1/accounts.
func (h *AccountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.createAccountUC.Execute(r.Context(), accountapp.CreateAccountInput{
		UserID:         userID,
		Name:           req.Name,
		Group:          req.Group,
		InitialBalance: req.InitialBalance,
		Currency:       req.Currency,
		Description:    req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainAccount.ErrInvalidAccountName):
			JSONError(w, http.StatusBadRequest, "INVALID_NAME", "Account name cannot be empty")
		case errors.Is(err, domainAccount.ErrInvalidGroup):
			JSONError(w, http.StatusBadRequest, "INVALID_GROUP", "Invalid account group")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create account")
		}
		return
	}

	JSON(w, http.StatusCreated, res, "Account created successfully")
}

// GetAccount handles GET /api/v1/accounts/{id}.
func (h *AccountHandler) GetAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	accountID := extractAccountID(r)
	if accountID == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Account ID is required")
		return
	}

	acc, err := h.getAccountUC.Execute(r.Context(), userID, accountID)
	if err != nil {
		if errors.Is(err, domainAccount.ErrAccountNotFound) {
			JSONError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve account")
		return
	}

	JSON(w, http.StatusOK, acc, "")
}

type UpdateAccountRequest struct {
	Name        *string `json:"name"`
	Group       *string `json:"group"`
	Description *string `json:"description"`
}

// UpdateAccount handles PATCH /api/v1/accounts/{id}.
func (h *AccountHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	accountID := extractAccountID(r)
	if accountID == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Account ID is required")
		return
	}

	var req UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.updateAccountUC.Execute(r.Context(), accountapp.UpdateAccountInput{
		UserID:      userID,
		AccountID:   accountID,
		Name:        req.Name,
		Group:       req.Group,
		Description: req.Description,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainAccount.ErrAccountNotFound):
			JSONError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
		case errors.Is(err, domainAccount.ErrInvalidAccountName):
			JSONError(w, http.StatusBadRequest, "INVALID_NAME", "Account name cannot be empty")
		case errors.Is(err, domainAccount.ErrInvalidGroup):
			JSONError(w, http.StatusBadRequest, "INVALID_GROUP", "Invalid account group")
		case errors.Is(err, domainAccount.ErrNoFieldsToUpdate):
			JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "At least one field must be provided to update")
		default:
			JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update account")
		}
		return
	}

	JSON(w, http.StatusOK, res, "Account updated successfully")
}

// DeleteAccount handles DELETE /api/v1/accounts/{id}.
func (h *AccountHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	accountID := extractAccountID(r)
	if accountID == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_ACCOUNT_ID", "Account ID is required")
		return
	}

	err := h.deleteAccountUC.Execute(r.Context(), userID, accountID)
	if err != nil {
		if errors.Is(err, domainAccount.ErrAccountNotFound) {
			JSONError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
			return
		}
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to delete account")
		return
	}

	JSON(w, http.StatusOK, nil, "Account deleted successfully")
}

// GetSummary handles GET /api/v1/accounts/summary.
func (h *AccountHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	summary, err := h.getSummaryUC.Execute(r.Context(), userID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve accounts summary")
		return
	}

	JSON(w, http.StatusOK, summary, "")
}

func extractAccountID(r *http.Request) string {
	id := r.PathValue("id")
	if id != "" {
		return id
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "accounts" {
		return parts[3]
	}
	return ""
}
