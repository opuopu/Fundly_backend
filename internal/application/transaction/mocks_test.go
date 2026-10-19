package transaction_test

import (
	"context"

	"github.com/Dodul01/fundly/internal/domain/account"
	domainAttachment "github.com/Dodul01/fundly/internal/domain/attachment"
	"github.com/Dodul01/fundly/internal/domain/category"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

// --- transaction repository mock -------------------------------------------------------

type mockTxRepo struct {
	txs       map[string]*domainTx.Transaction
	createErr error
	getErr    error
	updateErr error
	deleteErr error
	listErr   error
}

func newMockTxRepo() *mockTxRepo {
	return &mockTxRepo{txs: make(map[string]*domainTx.Transaction)}
}

func (m *mockTxRepo) Create(_ context.Context, tx *domainTx.Transaction) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.txs[tx.ID] = tx
	return nil
}

func (m *mockTxRepo) GetByID(_ context.Context, id string) (*domainTx.Transaction, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	tx, ok := m.txs[id]
	if !ok {
		return nil, domainTx.ErrTransactionNotFound
	}
	return tx, nil
}

func (m *mockTxRepo) ListByAccountID(_ context.Context, accountID string) ([]*domainTx.Transaction, error) {
	var out []*domainTx.Transaction
	for _, tx := range m.txs {
		if tx.FromAccountID == accountID || tx.ToAccountID == accountID {
			out = append(out, tx)
		}
	}
	return out, nil
}

func (m *mockTxRepo) List(_ context.Context, filter domainTx.Filter) ([]*domainTx.Transaction, int, error) {
	if m.listErr != nil {
		return nil, 0, m.listErr
	}
	var out []*domainTx.Transaction
	for _, tx := range m.txs {
		if filter.UserID != "" && tx.UserID != filter.UserID {
			continue
		}
		if filter.Type != "" && string(tx.Type) != filter.Type {
			continue
		}
		if filter.CategoryID != "" && tx.CategoryID != filter.CategoryID {
			continue
		}
		out = append(out, tx)
	}
	return out, len(out), nil
}

func (m *mockTxRepo) Update(_ context.Context, tx *domainTx.Transaction) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	if _, ok := m.txs[tx.ID]; !ok {
		return domainTx.ErrTransactionNotFound
	}
	m.txs[tx.ID] = tx
	return nil
}

func (m *mockTxRepo) Delete(_ context.Context, id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	if _, ok := m.txs[id]; !ok {
		return domainTx.ErrTransactionNotFound
	}
	delete(m.txs, id)
	return nil
}

func (m *mockTxRepo) Search(_ context.Context, userID, _ string, _ int) ([]*domainTx.Transaction, error) {
	var out []*domainTx.Transaction
	for _, tx := range m.txs {
		if tx.UserID == userID {
			out = append(out, tx)
		}
	}
	return out, nil
}

// --- account repository mock -----------------------------------------------------------

type mockAccountRepo struct {
	accounts map[string]*account.Account
}

func newMockAccountRepo() *mockAccountRepo {
	return &mockAccountRepo{accounts: make(map[string]*account.Account)}
}

func (m *mockAccountRepo) GetByID(_ context.Context, id string) (*account.Account, error) {
	acc, ok := m.accounts[id]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	return acc, nil
}

func (m *mockAccountRepo) GetByUserID(_ context.Context, userID string) ([]*account.Account, error) {
	var out []*account.Account
	for _, acc := range m.accounts {
		if acc.UserID == userID {
			out = append(out, acc)
		}
	}
	return out, nil
}

func (m *mockAccountRepo) GetByUserIDAndGroup(_ context.Context, _, _ string) ([]*account.Account, error) {
	return nil, nil
}

func (m *mockAccountRepo) Create(_ context.Context, acc *account.Account) error {
	m.accounts[acc.ID] = acc
	return nil
}

func (m *mockAccountRepo) Update(_ context.Context, acc *account.Account) error {
	m.accounts[acc.ID] = acc
	return nil
}

func (m *mockAccountRepo) Delete(_ context.Context, id string) error {
	delete(m.accounts, id)
	return nil
}

// --- category repository mock ----------------------------------------------------------

type mockCategoryRepo struct {
	categories map[string]*category.Category
}

func newMockCategoryRepo() *mockCategoryRepo {
	return &mockCategoryRepo{categories: make(map[string]*category.Category)}
}

func (m *mockCategoryRepo) GetAll(_ context.Context) ([]*category.Category, error) { return nil, nil }

func (m *mockCategoryRepo) GetByType(_ context.Context, _ category.CategoryType) ([]*category.Category, error) {
	return nil, nil
}

func (m *mockCategoryRepo) GetByID(_ context.Context, id string) (*category.Category, error) {
	cat, ok := m.categories[id]
	if !ok {
		return nil, category.ErrCategoryNotFound
	}
	return cat, nil
}

func (m *mockCategoryRepo) Create(_ context.Context, cat *category.Category) error {
	m.categories[cat.ID] = cat
	return nil
}

func (m *mockCategoryRepo) Update(_ context.Context, cat *category.Category) error {
	m.categories[cat.ID] = cat
	return nil
}

func (m *mockCategoryRepo) Delete(_ context.Context, id string) error {
	delete(m.categories, id)
	return nil
}

// --- attachment provider mock ----------------------------------------------------------

type mockAttachmentProvider struct {
	byTx   map[string][]*domainAttachment.Attachment
	linked map[string][]string // transactionID -> attachment IDs linked
}

func newMockAttachmentProvider() *mockAttachmentProvider {
	return &mockAttachmentProvider{
		byTx:   make(map[string][]*domainAttachment.Attachment),
		linked: make(map[string][]string),
	}
}

func (m *mockAttachmentProvider) ListByTransactionID(_ context.Context, txID string) ([]*domainAttachment.Attachment, error) {
	return m.byTx[txID], nil
}

func (m *mockAttachmentProvider) CountByTransactionID(_ context.Context, txID string) (int, error) {
	return len(m.byTx[txID]), nil
}

func (m *mockAttachmentProvider) LinkToTransaction(_ context.Context, _ string, txID string, ids []string) error {
	m.linked[txID] = append(m.linked[txID], ids...)
	for _, id := range ids {
		m.byTx[txID] = append(m.byTx[txID], &domainAttachment.Attachment{ID: id, FileName: id + ".png", FileURL: "/uploads/attachments/" + id})
	}
	return nil
}

// seedAccount is a small helper to register an account with a starting balance.
func seedAccount(repo *mockAccountRepo, id, userID string, balance int64) *account.Account {
	acc := account.New(id, userID, "USD")
	if balance > 0 {
		_ = acc.Deposit(balance)
	}
	repo.accounts[id] = acc
	return acc
}
