package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	txapp "github.com/Dodul01/fundly/internal/application/transaction"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

type TransactionHandler struct {
	listUC     *txapp.ListTransactionsUseCase
	createUC   *txapp.CreateTransactionUseCase
	getUC      *txapp.GetTransactionUseCase
	updateUC   *txapp.UpdateTransactionUseCase
	deleteUC   *txapp.DeleteTransactionUseCase
	searchUC   *txapp.SearchTransactionsUseCase
	transferUC *txapp.TransferUseCase
}

func NewTransactionHandler(
	listUC *txapp.ListTransactionsUseCase,
	createUC *txapp.CreateTransactionUseCase,
	getUC *txapp.GetTransactionUseCase,
	updateUC *txapp.UpdateTransactionUseCase,
	deleteUC *txapp.DeleteTransactionUseCase,
	searchUC *txapp.SearchTransactionsUseCase,
	transferUC *txapp.TransferUseCase,
) *TransactionHandler {
	return &TransactionHandler{
		listUC:     listUC,
		createUC:   createUC,
		getUC:      getUC,
		updateUC:   updateUC,
		deleteUC:   deleteUC,
		searchUC:   searchUC,
		transferUC: transferUC,
	}
}

// HandleTransactions routes /api/v1/transactions.
func (h *TransactionHandler) HandleTransactions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListTransactions(w, r)
	case http.MethodPost:
		h.CreateTransaction(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// HandleTransactionByID routes /api/v1/transactions/{id}.
func (h *TransactionHandler) HandleTransactionByID(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetTransaction(w, r)
	case http.MethodPatch:
		h.UpdateTransaction(w, r)
	case http.MethodDelete:
		h.DeleteTransaction(w, r)
	default:
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
	}
}

// ListTransactions handles GET /api/v1/transactions.
func (h *TransactionHandler) ListTransactions(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	items, meta, err := h.listUC.Execute(r.Context(), txapp.ListTransactionsInput{
		UserID:     userID,
		Type:       q.Get("type"),
		CategoryID: q.Get("category_id"),
		AccountID:  q.Get("account_id"),
		Date:       q.Get("date"),
		StartDate:  q.Get("start_date"),
		EndDate:    q.Get("end_date"),
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve transactions")
		return
	}

	JSONPaginated(w, http.StatusOK, items, Pagination{
		Page:         meta.Page,
		Limit:        meta.Limit,
		TotalRecords: meta.TotalRecords,
		TotalPages:   meta.TotalPages,
	})
}

type CreateTransactionRequest struct {
	Type          string   `json:"type"`
	Date          string   `json:"date"`
	Amount        int64    `json:"amount"`
	Currency      string   `json:"currency"`
	CategoryID    string   `json:"category_id"`
	AccountID     string   `json:"account_id"`
	Note          string   `json:"note"`
	Description   string   `json:"description"`
	AttachmentIDs []string `json:"attachment_ids"`
}

// CreateTransaction handles POST /api/v1/transactions.
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.createUC.Execute(r.Context(), txapp.CreateTransactionInput{
		UserID:        userID,
		Type:          req.Type,
		Date:          req.Date,
		Amount:        req.Amount,
		Currency:      req.Currency,
		CategoryID:    req.CategoryID,
		AccountID:     req.AccountID,
		Note:          req.Note,
		Description:   req.Description,
		AttachmentIDs: req.AttachmentIDs,
	})
	if err != nil {
		writeTransactionError(w, err, "Failed to create transaction")
		return
	}

	JSON(w, http.StatusCreated, res, "Transaction created successfully")
}

// GetTransaction handles GET /api/v1/transactions/{id}.
func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractTransactionID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "Transaction ID is required")
		return
	}

	res, err := h.getUC.Execute(r.Context(), userID, id)
	if err != nil {
		writeTransactionError(w, err, "Failed to retrieve transaction")
		return
	}

	JSON(w, http.StatusOK, res, "")
}

type UpdateTransactionRequest struct {
	Amount        *int64   `json:"amount"`
	Date          *string  `json:"date"`
	CategoryID    *string  `json:"category_id"`
	AccountID     *string  `json:"account_id"`
	Note          *string  `json:"note"`
	Description   *string  `json:"description"`
	AttachmentIDs []string `json:"attachment_ids"`
}

// UpdateTransaction handles PATCH /api/v1/transactions/{id}.
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractTransactionID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "Transaction ID is required")
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.updateUC.Execute(r.Context(), txapp.UpdateTransactionInput{
		UserID:        userID,
		ID:            id,
		Amount:        req.Amount,
		Date:          req.Date,
		CategoryID:    req.CategoryID,
		AccountID:     req.AccountID,
		Note:          req.Note,
		Description:   req.Description,
		AttachmentIDs: req.AttachmentIDs,
	})
	if err != nil {
		writeTransactionError(w, err, "Failed to update transaction")
		return
	}

	JSON(w, http.StatusOK, res, "Transaction updated successfully")
}

// DeleteTransaction handles DELETE /api/v1/transactions/{id}.
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	id := extractTransactionID(r)
	if id == "" {
		JSONError(w, http.StatusBadRequest, "INVALID_TRANSACTION_ID", "Transaction ID is required")
		return
	}

	if err := h.deleteUC.Execute(r.Context(), userID, id); err != nil {
		writeTransactionError(w, err, "Failed to delete transaction")
		return
	}

	JSON(w, http.StatusOK, nil, "Transaction deleted successfully")
}

// Search handles GET /api/v1/transactions/search.
func (h *TransactionHandler) Search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	items, err := h.searchUC.Execute(r.Context(), userID, r.URL.Query().Get("q"), limit)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to search transactions")
		return
	}

	JSON(w, http.StatusOK, items, "")
}

type TransferRequest struct {
	FromAccountID string `json:"from_account_id"`
	ToAccountID   string `json:"to_account_id"`
	Amount        int64  `json:"amount"`
	Note          string `json:"note"`
}

// Transfer handles POST /api/v1/transactions/transfer.
func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	userID := UserIDFromContext(r.Context())
	if userID == "" {
		JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authenticated user context")
		return
	}

	var req TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	res, err := h.transferUC.Execute(r.Context(), txapp.TransferInput{
		UserID:        userID,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Note:          req.Note,
	})
	if err != nil {
		writeTransactionError(w, err, "Failed to execute transfer")
		return
	}

	JSON(w, http.StatusCreated, res, "Transfer completed successfully")
}

// writeTransactionError maps domain errors to the standard error envelope.
func writeTransactionError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, domainTx.ErrTransactionNotFound):
		JSONError(w, http.StatusNotFound, "TRANSACTION_NOT_FOUND", "Transaction not found")
	case errors.Is(err, domainAccount.ErrAccountNotFound):
		JSONError(w, http.StatusNotFound, "ACCOUNT_NOT_FOUND", "Account not found")
	case errors.Is(err, domainAccount.ErrInsufficientBalance):
		JSONError(w, http.StatusUnprocessableEntity, "INSUFFICIENT_BALANCE", "Insufficient account balance")
	case errors.Is(err, domainTx.ErrInvalidAmount), errors.Is(err, domainAccount.ErrInvalidAmount):
		JSONError(w, http.StatusBadRequest, "INVALID_AMOUNT", "Amount must be greater than zero")
	case errors.Is(err, domainTx.ErrInvalidType):
		JSONError(w, http.StatusBadRequest, "INVALID_TYPE", "Invalid transaction type: must be income or expense")
	case errors.Is(err, domainTx.ErrAccountRequired):
		JSONError(w, http.StatusBadRequest, "ACCOUNT_REQUIRED", "account_id is required")
	case errors.Is(err, domainTx.ErrCategoryRequired):
		JSONError(w, http.StatusBadRequest, "CATEGORY_REQUIRED", "category_id is required")
	case errors.Is(err, domainTx.ErrNoFieldsToUpdate):
		JSONError(w, http.StatusBadRequest, "INVALID_REQUEST", "At least one field must be provided to update")
	default:
		JSONError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fallback)
	}
}

func extractTransactionID(r *http.Request) string {
	id := r.PathValue("id")
	if id != "" {
		return id
	}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 4 && parts[0] == "api" && parts[1] == "v1" && parts[2] == "transactions" {
		return parts[3]
	}
	return ""
}
