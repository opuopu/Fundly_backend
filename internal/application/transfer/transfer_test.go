package transfer_test

import (
	"context"
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

func (m *mockAccountRepo) Create(_ context.Context, acc *account.Account) error {
	m.accounts[acc.ID] = acc
	return nil
}

func (m *mockAccountRepo) Update(_ context.Context, acc *account.Account) error {
	m.accounts[acc.ID] = acc
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

func TestTransferServiceSuccess(t *testing.T) {
	accFrom := account.New("acc-1", "usr-1", "USD")
	_ = accFrom.Deposit(1000)

	accTo := account.New("acc-2", "usr-2", "USD")

	accRepo := &mockAccountRepo{
		accounts: map[string]*account.Account{
			"acc-1": accFrom,
			"acc-2": accTo,
		},
	}
	txRepo := &mockTxRepo{}

	service := transfer.NewService(accRepo, txRepo)

	req := transfer.Request{
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
}
