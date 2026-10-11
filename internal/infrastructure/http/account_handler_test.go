package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appaccount "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
	"github.com/Dodul01/fundly/internal/infrastructure/security"
)

type mockAccountRepoForHandler struct {
	accounts map[string]*domainAccount.Account
	err      error
}

func (m *mockAccountRepoForHandler) GetByID(ctx context.Context, id string) (*domainAccount.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	acc, ok := m.accounts[id]
	if !ok {
		return nil, domainAccount.ErrAccountNotFound
	}
	return acc, nil
}

func (m *mockAccountRepoForHandler) GetByUserID(ctx context.Context, userID string) ([]*domainAccount.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainAccount.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID {
			res = append(res, acc)
		}
	}
	return res, nil
}

func (m *mockAccountRepoForHandler) GetByUserIDAndGroup(ctx context.Context, userID, group string) ([]*domainAccount.Account, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainAccount.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID && strings.EqualFold(acc.Group, group) {
			res = append(res, acc)
		}
	}
	return res, nil
}

func (m *mockAccountRepoForHandler) Create(ctx context.Context, acc *domainAccount.Account) error {
	m.accounts[acc.ID] = acc
	return nil
}

func (m *mockAccountRepoForHandler) Update(ctx context.Context, acc *domainAccount.Account) error {
	m.accounts[acc.ID] = acc
	return nil
}

func (m *mockAccountRepoForHandler) Delete(ctx context.Context, id string) error {
	delete(m.accounts, id)
	return nil
}

func setupTestAccountHandler(repo *mockAccountRepoForHandler) *infrahttp.AccountHandler {
	listUC := appaccount.NewListAccountsUseCase(repo)
	createUC := appaccount.NewCreateAccountUseCase(repo)
	getUC := appaccount.NewGetAccountUseCase(repo)
	updateUC := appaccount.NewUpdateAccountUseCase(repo)
	deleteUC := appaccount.NewDeleteAccountUseCase(repo)
	summaryUC := appaccount.NewGetAccountsSummaryUseCase(repo)
	return infrahttp.NewAccountHandler(listUC, createUC, getUC, updateUC, deleteUC, summaryUC)
}

func TestAccountHandler_ListAccounts_All(t *testing.T) {
	repo := &mockAccountRepoForHandler{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc1, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Daily cash")
	acc2, _ := domainAccount.NewAccount("acc_02", "usr_1", "Debit Card", "card", "USD", 154050, "Checking card")
	repo.accounts[acc1.ID] = acc1
	repo.accounts[acc2.ID] = acc2

	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                         `json:"success"`
		Data    []appaccount.AccountResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if len(resp.Data) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(resp.Data))
	}
	if resp.Data[0].Name != "Cash" && resp.Data[1].Name != "Cash" {
		t.Errorf("expected 'Cash' account in list")
	}
}

func TestAccountHandler_ListAccounts_WithGroupFilter(t *testing.T) {
	repo := &mockAccountRepoForHandler{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc1, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Daily cash")
	acc2, _ := domainAccount.NewAccount("acc_02", "usr_1", "Debit Card", "card", "USD", 154050, "Checking card")
	repo.accounts[acc1.ID] = acc1
	repo.accounts[acc2.ID] = acc2

	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts?group=card", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                         `json:"success"`
		Data    []appaccount.AccountResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 card account, got %d", len(resp.Data))
	}
	if resp.Data[0].Group != "card" {
		t.Errorf("expected group 'card', got '%s'", resp.Data[0].Group)
	}
}

func TestAccountHandler_ListAccounts_Empty(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_empty")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                         `json:"success"`
		Data    []appaccount.AccountResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Data == nil {
		t.Errorf("expected non-nil empty list in data")
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0 accounts, got %d", len(resp.Data))
	}
}

func TestAccountHandler_ListAccounts_Unauthenticated(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAccountHandler_MethodNotAllowed(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestAccountHandler_WithAuthMiddleware(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	jwtSvc := security.NewJWTService("secret-accounts-key", time.Hour, 24*time.Hour)
	blacklist := security.NewMemoryTokenBlacklist()
	authMiddleware := infrahttp.NewAuthMiddleware(jwtSvc, blacklist)

	protectedHandler := authMiddleware.RequireAuth(handler.HandleAccounts)

	// Missing token
	reqNoAuth := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	recNoAuth := httptest.NewRecorder()
	protectedHandler(recNoAuth, reqNoAuth)
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", recNoAuth.Code)
	}

	// Valid token
	tokens, _ := jwtSvc.GenerateTokenPair("usr_1", "user@fundly.com")
	reqValid := httptest.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	reqValid.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	recValid := httptest.NewRecorder()
	protectedHandler(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recValid.Code)
	}
}

func TestAccountHandler_CreateAccount_Success(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	body := `{"group":"cash","name":"Cash Wallet","initial_balance":20000,"currency":"USD","description":"Physical wallet cash"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    appaccount.AccountResponse `json:"data"`
		Message string                     `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Data.Name != "Cash Wallet" {
		t.Errorf("expected name 'Cash Wallet', got '%s'", resp.Data.Name)
	}
	if resp.Data.Group != "cash" {
		t.Errorf("expected group 'cash', got '%s'", resp.Data.Group)
	}
	if resp.Data.Balance != 20000 {
		t.Errorf("expected balance 20000, got %d", resp.Data.Balance)
	}
	if resp.Data.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", resp.Data.Currency)
	}
	if resp.Data.Description != "Physical wallet cash" {
		t.Errorf("expected description 'Physical wallet cash', got '%s'", resp.Data.Description)
	}
	if resp.Data.ID == "" {
		t.Errorf("expected generated account ID")
	}
}

func TestAccountHandler_CreateAccount_InvalidJSON(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader("invalid-json{"))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestAccountHandler_CreateAccount_EmptyName(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	body := `{"group":"cash","name":"","initial_balance":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestAccountHandler_CreateAccount_InvalidGroup(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	body := `{"group":"unknown_group","name":"Test","initial_balance":0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request, got %d", rec.Code)
	}
}

func TestAccountHandler_CreateAccount_Unauthenticated(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	body := `{"group":"cash","name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.HandleAccounts(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", rec.Code)
	}
}

func TestAccountHandler_GetAccount_Success(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Primary Wallet", "cash", "USD", 50000, "Daily cash")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/acc_01", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    appaccount.AccountResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Data.ID != "acc_01" {
		t.Errorf("expected ID 'acc_01', got '%s'", resp.Data.ID)
	}
	if resp.Data.Name != "Primary Wallet" {
		t.Errorf("expected name 'Primary Wallet', got '%s'", resp.Data.Name)
	}
	if resp.Data.Balance != 50000 {
		t.Errorf("expected balance 50000, got %d", resp.Data.Balance)
	}
}

func TestAccountHandler_GetAccount_NotFound(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/non_existent", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestAccountHandler_GetAccount_OtherUserForbidden(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_02", "usr_2", "Secret Fund", "savings", "USD", 1000000, "")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	// usr_1 tries to access usr_2's account
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/acc_02", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestAccountHandler_GetAccount_Unauthenticated(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/acc_01", nil)
	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAccountHandler_HandleAccountByID_MethodNotAllowed(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/acc_01", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestAccountHandler_UpdateAccount_Success(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	body := `{"name":"Primary Cash Wallet","group":"card","description":"Updated description"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/acc_01", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    appaccount.AccountResponse `json:"data"`
		Message string                     `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Data.Name != "Primary Cash Wallet" {
		t.Errorf("expected name 'Primary Cash Wallet', got '%s'", resp.Data.Name)
	}
	if resp.Data.Group != "card" {
		t.Errorf("expected group 'card', got '%s'", resp.Data.Group)
	}
	if resp.Data.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got '%s'", resp.Data.Description)
	}
}

func TestAccountHandler_UpdateAccount_PartialUpdate(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	body := `{"name":"New Cash Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/acc_01", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    appaccount.AccountResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.Name != "New Cash Name" || resp.Data.Group != "cash" {
		t.Errorf("unexpected fields: %+v", resp.Data)
	}
}

func TestAccountHandler_UpdateAccount_InvalidJSON(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/acc_01", strings.NewReader("invalid-json{"))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestAccountHandler_UpdateAccount_EmptyUpdate(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	body := `{}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/acc_01", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestAccountHandler_UpdateAccount_InvalidName(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	body := `{"name":"   "}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/acc_01", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestAccountHandler_UpdateAccount_InvalidGroup(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	body := `{"group":"crypto"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/acc_01", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestAccountHandler_UpdateAccount_NotFound(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	body := `{"name":"Test"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/non_existent", strings.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestAccountHandler_UpdateAccount_Unauthenticated(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	body := `{"name":"Test"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/accounts/acc_01", strings.NewReader(body))
	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAccountHandler_DeleteAccount_Success(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/acc_01", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Message != "Account deleted successfully" {
		t.Errorf("expected message 'Account deleted successfully', got '%s'", resp.Message)
	}
	if _, exists := repo.accounts["acc_01"]; exists {
		t.Errorf("expected account to be deleted from repo")
	}
}

func TestAccountHandler_DeleteAccount_NotFound(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/non_existent", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestAccountHandler_DeleteAccount_ForbiddenOtherUser(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_02", "usr_2", "Secret", "savings", "USD", 5000, "")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	// usr_1 tries to delete usr_2's account
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/acc_02", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
	if _, exists := repo.accounts["acc_02"]; !exists {
		t.Errorf("other user's account should not be deleted")
	}
}

func TestAccountHandler_DeleteAccount_Unauthenticated(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/accounts/acc_01", nil)
	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAccountHandler_GetSummary_Success(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc1, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "")
	acc2, _ := domainAccount.NewAccount("acc_02", "usr_1", "Card", "card", "USD", 154050, "")
	acc3, _ := domainAccount.NewAccount("acc_03", "usr_1", "Savings", "savings", "USD", 500000, "")
	repo.accounts[acc1.ID] = acc1
	repo.accounts[acc2.ID] = acc2
	repo.accounts[acc3.ID] = acc3
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/summary", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                                `json:"success"`
		Data    appaccount.AccountsSummaryResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true, got false")
	}
	if resp.Data.TotalBalance != (20000 + 154050 + 500000) {
		t.Errorf("expected total balance %d, got %d", (20000 + 154050 + 500000), resp.Data.TotalBalance)
	}
	if resp.Data.TotalCash != 20000 {
		t.Errorf("expected total cash 20000, got %d", resp.Data.TotalCash)
	}
	if resp.Data.TotalCard != 154050 {
		t.Errorf("expected total card 154050, got %d", resp.Data.TotalCard)
	}
	if resp.Data.TotalSavings != 500000 {
		t.Errorf("expected total savings 500000, got %d", resp.Data.TotalSavings)
	}
	if resp.Data.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", resp.Data.Currency)
	}
}

func TestAccountHandler_GetSummary_Empty(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/summary", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetSummary(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                                `json:"success"`
		Data    appaccount.AccountsSummaryResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Data.TotalBalance != 0 || resp.Data.TotalCash != 0 {
		t.Errorf("expected all 0 balances for empty account list, got %+v", resp.Data)
	}
}

func TestAccountHandler_GetSummary_ViaHandleAccountByID(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "")
	repo.accounts[acc.ID] = acc
	handler := setupTestAccountHandler(repo)

	// Routing to /api/v1/accounts/summary through HandleAccountByID dispatcher
	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/summary", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleAccountByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestAccountHandler_GetSummary_Unauthenticated(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/accounts/summary", nil)
	rec := httptest.NewRecorder()
	handler.GetSummary(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAccountHandler_GetSummary_MethodNotAllowed(t *testing.T) {
	repo := &mockAccountRepoForHandler{accounts: make(map[string]*domainAccount.Account)}
	handler := setupTestAccountHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/accounts/summary", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.GetSummary(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}





