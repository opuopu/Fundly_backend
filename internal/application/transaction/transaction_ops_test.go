package transaction_test

import (
	"context"
	"errors"
	"testing"

	txapp "github.com/Dodul01/fundly/internal/application/transaction"
	domainTx "github.com/Dodul01/fundly/internal/domain/transaction"
)

// helper to create one expense entry owned by usr_1 and return its ID.
func seedExpense(t *testing.T, txRepo *mockTxRepo, accRepo *mockAccountRepo, catRepo *mockCategoryRepo) string {
	t.Helper()
	seedAccount(accRepo, "acc_1", "usr_1", 10000)
	uc := txapp.NewCreateTransactionUseCase(txRepo, accRepo, catRepo, nil, nil)
	res, err := uc.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:     "usr_1",
		Type:       "expense",
		Amount:     2000,
		CategoryID: "cat_food",
		AccountID:  "acc_1",
	})
	if err != nil {
		t.Fatalf("seed expense failed: %v", err)
	}
	return res.ID
}

func TestListTransactions_Pagination(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	seedExpense(t, txRepo, accRepo, catRepo)

	uc := txapp.NewListTransactionsUseCase(txRepo, accRepo, catRepo, nil)
	items, meta, err := uc.Execute(context.Background(), txapp.ListTransactionsInput{UserID: "usr_1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if meta.Page != 1 || meta.Limit != 20 || meta.TotalRecords != 1 || meta.TotalPages != 1 {
		t.Errorf("unexpected pagination meta: %+v", meta)
	}
}

func TestGetTransaction_OwnershipGuard(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	id := seedExpense(t, txRepo, accRepo, catRepo)

	uc := txapp.NewGetTransactionUseCase(txRepo, accRepo, catRepo, nil)

	if _, err := uc.Execute(context.Background(), "usr_1", id); err != nil {
		t.Fatalf("owner should fetch: %v", err)
	}

	_, err := uc.Execute(context.Background(), "intruder", id)
	if !errors.Is(err, domainTx.ErrTransactionNotFound) {
		t.Fatalf("expected not-found for non-owner, got %v", err)
	}
}

func TestUpdateTransaction_AdjustsBalance(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	id := seedExpense(t, txRepo, accRepo, catRepo) // acc_1 now 8000 (10000 - 2000)

	uc := txapp.NewUpdateTransactionUseCase(txRepo, accRepo, catRepo, nil, nil)
	newAmount := int64(5000)
	_, err := uc.Execute(context.Background(), txapp.UpdateTransactionInput{
		UserID: "usr_1",
		ID:     id,
		Amount: &newAmount,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Reverse old (+2000 => 10000), apply new (-5000 => 5000).
	if got := accRepo.accounts["acc_1"].Balance; got != 5000 {
		t.Errorf("expected balance 5000 after amount change, got %d", got)
	}
}

func TestUpdateTransaction_LinksAttachments(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	att := newMockAttachmentProvider()
	seedAccount(accRepo, "acc_1", "usr_1", 10000)

	createUC := txapp.NewCreateTransactionUseCase(txRepo, accRepo, catRepo, att, nil)
	created, err := createUC.Execute(context.Background(), txapp.CreateTransactionInput{
		UserID:        "usr_1",
		Type:          "expense",
		Amount:        2000,
		CategoryID:    "cat_food",
		AccountID:     "acc_1",
		AttachmentIDs: []string{"att_1"},
	})
	if err != nil {
		t.Fatalf("seed transaction failed: %v", err)
	}

	uc := txapp.NewUpdateTransactionUseCase(txRepo, accRepo, catRepo, att, nil)
	res, err := uc.Execute(context.Background(), txapp.UpdateTransactionInput{
		UserID:        "usr_1",
		ID:            created.ID,
		AttachmentIDs: []string{"att_2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.AttachmentCount != 2 {
		t.Errorf("expected attachment_count 2 after update, got %d", res.AttachmentCount)
	}
	if len(att.linked[created.ID]) != 2 {
		t.Errorf("expected 2 attachments linked to tx, got %d: %v", len(att.linked[created.ID]), att.linked[created.ID])
	}
}

func TestDeleteTransaction_ReversesBalance(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	id := seedExpense(t, txRepo, accRepo, catRepo) // acc_1 at 8000

	uc := txapp.NewDeleteTransactionUseCase(txRepo, accRepo, nil)
	if err := uc.Execute(context.Background(), "usr_1", id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := accRepo.accounts["acc_1"].Balance; got != 10000 {
		t.Errorf("expected balance restored to 10000, got %d", got)
	}
	if len(txRepo.txs) != 0 {
		t.Errorf("expected transaction removed, got %d", len(txRepo.txs))
	}
}

func TestDeleteTransaction_OwnershipGuard(t *testing.T) {
	txRepo := newMockTxRepo()
	accRepo := newMockAccountRepo()
	catRepo := newMockCategoryRepo()
	id := seedExpense(t, txRepo, accRepo, catRepo)

	uc := txapp.NewDeleteTransactionUseCase(txRepo, accRepo, nil)
	err := uc.Execute(context.Background(), "intruder", id)
	if !errors.Is(err, domainTx.ErrTransactionNotFound) {
		t.Fatalf("expected not-found for non-owner, got %v", err)
	}
}

func TestSearchTransactions_EmptyQuery(t *testing.T) {
	uc := txapp.NewSearchTransactionsUseCase(newMockTxRepo(), newMockAccountRepo(), newMockCategoryRepo(), nil)
	items, err := uc.Execute(context.Background(), "usr_1", "   ", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("expected empty result for blank query, got %d", len(items))
	}
}
