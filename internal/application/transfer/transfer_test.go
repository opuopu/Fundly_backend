package transfer_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dodul01/fundly/internal/application/transfer"
	"github.com/Dodul01/fundly/internal/domain/account"
	"github.com/Dodul01/fundly/internal/domain/transaction"
)

type mockAccountRepo struct {
	accounts map[string]*account.Account
}

func (m *mockAccountRepo) GetByID(_ context.Context, id string) (*account.Account, error) {
	acc, ok := m.accounts[id]
	if !ok {
		return nil, account.ErrAccountNotFound
	}
	return acc, nil
}

func (m *mockAccountRepo) GetByUserID(_ context.Context, _ string) ([]*account.Account, error) {
	return nil, nil
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

type mockTxRepo struct {
	txs []*transaction.Transaction
}

func (m *mockTxRepo) Create(_ context.Context, tx *transaction.Transaction) error {
	m.txs = append(m.txs, tx)
	return nil
}

func (m *mockTxRepo) GetByID(_ context.Context, _ string) (*transaction.Transaction, error) {
	return nil, nil
}

func (m *mockTxRepo) ListByAccountID(_ context.Context, _ string) ([]*transaction.Transaction, error) {
	return nil, nil
}

func (m *mockTxRepo) List(_ context.Context, _ transaction.Filter) ([]*transaction.Transaction, int, error) {
	return nil, 0, nil
}

func (m *mockTxRepo) Update(_ context.Context, _ *transaction.Transaction) error {
	return nil
}

func (m *mockTxRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockTxRepo) Search(_ context.Context, _, _ string, _ int) ([]*transaction.Transaction, error) {
	return nil, nil
}

func TestTransferServiceSuccess(t *testing.T) {
	accFrom := account.New("acc-1", "usr-1", "USD")
	_ = accFrom.Deposit(1000)

	accTo := account.New("acc-2", "usr-1", "USD")

	accRepo := &mockAccountRepo{
		accounts: map[string]*account.Account{
			"acc-1": accFrom,
			"acc-2": accTo,
		},
	}
	txRepo := &mockTxRepo{}

	service := transfer.NewService(accRepo, txRepo, nil)

	req := transfer.Request{
		UserID:        "usr-1",
		FromAccountID: "acc-1",
		ToAccountID:   "acc-2",
		Amount:        300,
	}

	tx, err := service.Execute(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error executing transfer: %v", err)
	}

	if tx.Status != transaction.StatusCompleted {
		t.Errorf("expected status COMPLETED, got %v", tx.Status)
	}
	if accFrom.Balance != 700 {
		t.Errorf("expected from balance 700, got %d", accFrom.Balance)
	}
	if accTo.Balance != 300 {
		t.Errorf("expected to balance 300, got %d", accTo.Balance)
	}
	if len(txRepo.txs) != 1 {
		t.Errorf("expected 1 recorded transaction, got %d", len(txRepo.txs))
	}
	if txRepo.txs[0].UserID != "usr-1" {
		t.Errorf("expected recorded transfer to carry owner usr-1, got %q", txRepo.txs[0].UserID)
	}
}

func TestTransferService_RejectsForeignAccount(t *testing.T) {
	accFrom := account.New("acc-1", "victim", "USD")
	_ = accFrom.Deposit(1000)
	accTo := account.New("acc-2", "attacker", "USD")

	accRepo := &mockAccountRepo{
		accounts: map[string]*account.Account{
			"acc-1": accFrom,
			"acc-2": accTo,
		},
	}
	txRepo := &mockTxRepo{}
	service := transfer.NewService(accRepo, txRepo, nil)

	// Attacker tries to pull funds out of the victim's account into their own.
	_, err := service.Execute(context.Background(), transfer.Request{
		UserID:        "attacker",
		FromAccountID: "acc-1",
		ToAccountID:   "acc-2",
		Amount:        300,
	})
	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound for foreign account, got %v", err)
	}
	if accFrom.Balance != 1000 {
		t.Errorf("victim balance must be untouched, got %d", accFrom.Balance)
	}
	if len(txRepo.txs) != 0 {
		t.Errorf("no transaction should be recorded on rejection, got %d", len(txRepo.txs))
	}
}
