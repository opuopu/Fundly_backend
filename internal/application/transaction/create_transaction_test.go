package transaction_test

import (
	"context"
	"errors"
	"testing"

	txapp "github.com/Dodul01/fundly/internal/application/transaction"
	"github.com/Dodul01/fundly/internal/domain/account"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

func TestCreateTransaction_Expense_DebitsAccount(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	seedAccount(accRepo, "acc_1", "usr_1", 10000)

	uc := txapp.NewCreateTransactionUseCase(txRepo, accRepo, catRepo, nil, nil)

	res, err := uc.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:     "usr_1",
		Type:       "expense",
		Amount:     3000,
		Currency:   "USD",
		CategoryID: "cat_food",
		AccountID:  "acc_1",
		Note:       "Lunch",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID == "" {
		t.Error("expected generated transaction ID")
	}
	if res.Type != "EXPENSE" {
		t.Errorf("expected type EXPENSE, got %s", res.Type)
	}
	if got := accRepo.accounts["acc_1"].Balance; got != 7000 {
		t.Errorf("expected balance 7000 after expense, got %d", got)
	}
	if len(txRepo.txs) != 1 {
		t.Errorf("expected transaction to be persisted, got %d", len(txRepo.txs))
	}
}

func TestCreateTransaction_Income_CreditsAccount(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	seedAccount(accRepo, "acc_1", "usr_1", 1000)

	uc := txapp.NewCreateTransactionUseCase(txRepo, accRepo, catRepo, nil, nil)

	_, err := uc.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:     "usr_1",
		Type:       "income",
		Amount:     5000,
		Currency:   "USD",
		CategoryID: "cat_salary",
		AccountID:  "acc_1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := accRepo.accounts["acc_1"].Balance; got != 6000 {
		t.Errorf("expected balance 6000 after income, got %d", got)
	}
}

func TestCreateTransaction_InsufficientBalance(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	seedAccount(accRepo, "acc_1", "usr_1", 1000)

	uc := txapp.NewCreateTransactionUseCase(txRepo, accRepo, catRepo, nil, nil)

	_, err := uc.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:     "usr_1",
		Type:       "expense",
		Amount:     5000,
		CategoryID: "cat_food",
		AccountID:  "acc_1",
	})
	if !errors.Is(err, account.ErrInsufficientBalance) {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
	if len(txRepo.txs) != 0 {
		t.Error("expected no transaction persisted on failed expense")
	}
}

func TestCreateTransaction_InvalidType(t *testing.T) {
	uc := txapp.NewCreateTransactionUseCase(newMockTxRepo(), newMockAccountRepo(), newMockCategoryRepo(), nil, nil)
	_, err := uc.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:     "usr_1",
		Type:       "TRANSFER",
		Amount:     100,
		CategoryID: "cat_food",
		AccountID:  "acc_1",
	})
	if !errors.Is(err, domainTx.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestCreateTransaction_LinksAttachments(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	att := newMockAttachmentProvider()
	seedAccount(accRepo, "acc_1", "usr_1", 10000)

	uc := txapp.NewCreateTransactionUseCase(txRepo, accRepo, catRepo, att, nil)
	res, err := uc.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:        "usr_1",
		Type:          "expense",
		Amount:        2000,
		CategoryID:    "cat_food",
		AccountID:     "acc_1",
		AttachmentIDs: []string{"att_1", "att_2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.AttachmentCount != 2 {
		t.Errorf("expected attachment_count 2, got %d", res.AttachmentCount)
	}
	if len(res.Attachments) != 2 {
		t.Errorf("expected 2 attachments in detail response, got %d", len(res.Attachments))
	}
	if len(att.linked[res.ID]) != 2 {
		t.Errorf("expected 2 attachments linked to tx, got %d", len(att.linked[res.ID]))
	}
}

func TestCreateTransaction_ForeignAccount(t *testing.T) {
	accRepo := newMockAccountRepo()
	seedAccount(accRepo, "acc_1", "other_user", 10000)

	uc := txapp.NewCreateTransactionUseCase(newMockTxRepo(), accRepo, newMockCategoryRepo(), nil, nil)
	_, err := uc.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:     "usr_1",
		Type:       "expense",
		Amount:     1000,
		CategoryID: "cat_food",
		AccountID:  "acc_1",
	})
	if !errors.Is(err, account.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound for foreign account, got %v", err)
	}
}
