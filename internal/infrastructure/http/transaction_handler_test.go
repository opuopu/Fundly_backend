package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	txapp "github.com/Dodul01/fundly/internal/application/transaction"
	"github.com/Dodul01/fundly/internal/application/transfer"
	"github.com/Dodul01/fundly/internal/domain/account"
	"github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

// --- mocks ----------------------------------------------------------------------------

type txRepoStub struct {
	txs map[string]*domainTx.Transaction
}

func (m *txRepoStub) Create(_ context.Context, tx *domainTx.Transaction) error {
	m.txs[tx.ID] = tx
	return nil
}
func (m *txRepoStub) GetByID(_ context.Context, id string) (*domainTx.Transaction, error) {
	tx, ok := m.txs[id]
	if !ok {
		return nil, domainTx.ErrTransactionNotFound
	}
	return tx, nil
}
func (m *txRepoStub) ListByAccountID(_ context.Context, _ string) ([]*domainTx.Transaction, error) {
	return nil, nil
}
func (m *txRepoStub) List(_ context.Context, f domainTx.Filter) ([]*domainTx.Transaction, int, error) {
	var out []*domainTx.Transaction
	for _, tx := range m.txs {
		if f.UserID == "" || tx.UserID == f.UserID {
			out = append(out, tx)
		}
	}
	return out, len(out), nil
}
func (m *txRepoStub) Update(_ context.Context, tx *domainTx.Transaction) error {
	m.txs[tx.ID] = tx
	return nil
}
func (m *txRepoStub) Delete(_ context.Context, id string) error {
	delete(m.txs, id)
	return nil
}
func (m *txRepoStub) Search(_ context.Context, userID, _ string, _ int) ([]*domainTx.Transaction, error) {
	var out []*domainTx.Transaction
	for _, tx := range m.txs {
		if tx.UserID == userID {
			out = append(out, tx)
		}
	}
	return out, nil
}

type accRepoStub struct {
	accounts map[string]*account.Account
}

func (m *accRepoStub) GetByID(_ context.Context, id string) (*account.Account, error) {
	acc, ok := m.accounts[id]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	return acc, nil
}
func (m *accRepoStub) GetByUserID(_ context.Context, _ string) ([]*account.Account, error) {
	return nil, nil
}
func (m *accRepoStub) GetByUserIDAndGroup(_ context.Context, _, _ string) ([]*account.Account, error) {
	return nil, nil
}
func (m *accRepoStub) Create(_ context.Context, acc *account.Account) error {
	m.accounts[acc.ID] = acc
	return nil
}
func (m *accRepoStub) Update(_ context.Context, acc *account.Account) error {
	m.accounts[acc.ID] = acc
	return nil
}
func (m *accRepoStub) Delete(_ context.Context, id string) error {
	delete(m.accounts, id)
	return nil
}

type catRepoStub struct{}

func (catRepoStub) GetAll(_ context.Context) ([]*category.Category, error) { return nil, nil }
func (catRepoStub) GetByType(_ context.Context, _ category.CategoryType) ([]*category.Category, error) {
	return nil, nil
}
func (catRepoStub) GetByID(_ context.Context, _ string) (*category.Category, error) {
	return nil, category.ErrCategoryNotFound
}
func (catRepoStub) Create(_ context.Context, _ *category.Category) error { return nil }
func (catRepoStub) Update(_ context.Context, _ *category.Category) error { return nil }
func (catRepoStub) Delete(_ context.Context, _ string) error            { return nil }

func setupTxHandler() (*infrahttp.TransactionHandler, *accRepoStub) {
	txRepo := &txRepoStub{txs: make(map[string]*domainTx.Transaction)}
	accRepo := &accRepoStub{accounts: make(map[string]*account.Account)}
	catRepo := catRepoStub{}

	transferSvc := transfer.NewService(accRepo, txRepo, nil)

	return infrahttp.NewTransactionHandler(
		txapp.NewListTransactionsUseCase(txRepo, accRepo, catRepo, nil),
		txapp.NewCreateTransactionUseCase(txRepo, accRepo, catRepo, nil, nil),
		txapp.NewGetTransactionUseCase(txRepo, accRepo, catRepo, nil),
		txapp.NewUpdateTransactionUseCase(txRepo, accRepo, catRepo, nil, nil),
		txapp.NewDeleteTransactionUseCase(txRepo, accRepo, nil),
		txapp.NewSearchTransactionsUseCase(txRepo, accRepo, catRepo, nil),
		txapp.NewTransferUseCase(transferSvc, accRepo, catRepo),
	), accRepo
}

func authed(req *http.Request, userID string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), infrahttp.UserIDKey, userID))
}

// --- tests ----------------------------------------------------------------------------

func TestTransactionHandler_Create(t *testing.T) {
	handler, accRepo := setupTxHandler()
	acc := account.New("acc_1", "usr_1", "USD")
	_ = acc.Deposit(10000)
	accRepo.accounts["acc_1"] = acc

	body := `{"type":"expense","amount":2500,"currency":"USD","category_id":"cat_food","account_id":"acc_1","note":"Coffee"}`
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(body)), "usr_1")
	rec := httptest.NewRecorder()

	handler.HandleTransactions(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. body: %s", rec.Code, rec.Body.String())
	}
	if accRepo.accounts["acc_1"].Balance != 7500 {
		t.Errorf("expected balance 7500, got %d", accRepo.accounts["acc_1"].Balance)
	}
}

func TestTransactionHandler_List_Paginated(t *testing.T) {
	handler, accRepo := setupTxHandler()
	acc := account.New("acc_1", "usr_1", "USD")
	_ = acc.Deposit(10000)
	accRepo.accounts["acc_1"] = acc

	createBody := `{"type":"income","amount":1000,"category_id":"cat_salary","account_id":"acc_1"}`
	createReq := authed(httptest.NewRequest(http.MethodPost, "/api/v1/transactions", strings.NewReader(createBody)), "usr_1")
	handler.HandleTransactions(httptest.NewRecorder(), createReq)

	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil), "usr_1")
	rec := httptest.NewRecorder()
	handler.HandleTransactions(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Success    bool `json:"success"`
		Pagination struct {
			TotalRecords int `json:"total_records"`
			TotalPages   int `json:"total_pages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if !resp.Success || resp.Pagination.TotalRecords != 1 || resp.Pagination.TotalPages != 1 {
		t.Errorf("unexpected pagination envelope: %s", rec.Body.String())
	}
}

func TestTransactionHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupTxHandler()
	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/transactions/tx_missing", nil), "usr_1")
	req.SetPathValue("id", "tx_missing")
	rec := httptest.NewRecorder()

	handler.HandleTransactionByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTransactionHandler_Unauthorized(t *testing.T) {
	handler, _ := setupTxHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions", nil) // no user context
	rec := httptest.NewRecorder()

	handler.HandleTransactions(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestTransactionHandler_Transfer(t *testing.T) {
	handler, accRepo := setupTxHandler()
	from := account.New("acc_1", "usr_1", "USD")
	_ = from.Deposit(10000)
	to := account.New("acc_2", "usr_1", "USD")
	accRepo.accounts["acc_1"] = from
	accRepo.accounts["acc_2"] = to

	body := `{"from_account_id":"acc_1","to_account_id":"acc_2","amount":4000,"note":"move"}`
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/transactions/transfer", strings.NewReader(body)), "usr_1")
	rec := httptest.NewRecorder()

	handler.Transfer(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. body: %s", rec.Code, rec.Body.String())
	}
	if accRepo.accounts["acc_1"].Balance != 6000 || accRepo.accounts["acc_2"].Balance != 4000 {
		t.Errorf("unexpected balances: from=%d to=%d", accRepo.accounts["acc_1"].Balance, accRepo.accounts["acc_2"].Balance)
	}

	var resp struct {
		Data txapp.TransferResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if resp.Data.FromAccount == nil || resp.Data.FromAccount.ID != "acc_1" {
		t.Errorf("expected from_account acc_1, got %+v", resp.Data.FromAccount)
	}
	if resp.Data.ToAccount == nil || resp.Data.ToAccount.ID != "acc_2" {
		t.Errorf("expected to_account acc_2, got %+v", resp.Data.ToAccount)
	}
}
