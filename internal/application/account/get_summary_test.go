package account_test

import (
	"context"
	"errors"
	"testing"

	appaccount "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

func TestGetAccountsSummary_Success(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc1, _ := domainAccount.NewAccount("acc_01", "usr_1", "Wallet", "cash", "USD", 20000, "")
	acc2, _ := domainAccount.NewAccount("acc_02", "usr_1", "Visa Card", "card", "USD", 154050, "")
	acc3, _ := domainAccount.NewAccount("acc_03", "usr_1", "Rainy Day", "savings", "USD", 500000, "")
	acc4, _ := domainAccount.NewAccount("acc_04", "usr_1", "Checking Acc", "checking", "USD", 50000, "")
	// Other user account that must not be counted
	accOther, _ := domainAccount.NewAccount("acc_other", "usr_2", "Other Wallet", "cash", "USD", 99999, "")

	repo.accounts[acc1.ID] = acc1
	repo.accounts[acc2.ID] = acc2
	repo.accounts[acc3.ID] = acc3
	repo.accounts[acc4.ID] = acc4
	repo.accounts[accOther.ID] = accOther

	uc := appaccount.NewGetAccountsSummaryUseCase(repo)

	summary, err := uc.Execute(context.Background(), "usr_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedTotal := int64(20000 + 154050 + 500000 + 50000)
	if summary.TotalBalance != expectedTotal {
		t.Errorf("expected total balance %d, got %d", expectedTotal, summary.TotalBalance)
	}
	if summary.TotalCash != 20000 {
		t.Errorf("expected total cash 20000, got %d", summary.TotalCash)
	}
	// card + checking
	if summary.TotalCard != (154050 + 50000) {
		t.Errorf("expected total card %d, got %d", (154050 + 50000), summary.TotalCard)
	}
	if summary.TotalSavings != 500000 {
		t.Errorf("expected total savings 500000, got %d", summary.TotalSavings)
	}
	if summary.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", summary.Currency)
	}
}

func TestGetAccountsSummary_NoAccounts(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewGetAccountsSummaryUseCase(repo)

	summary, err := uc.Execute(context.Background(), "usr_new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.TotalBalance != 0 {
		t.Errorf("expected total balance 0, got %d", summary.TotalBalance)
	}
	if summary.TotalCash != 0 || summary.TotalCard != 0 || summary.TotalSavings != 0 {
		t.Errorf("expected all category totals to be 0")
	}
	if summary.Currency != "USD" {
		t.Errorf("expected default currency 'USD', got '%s'", summary.Currency)
	}
}

func TestGetAccountsSummary_EmptyUserID(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewGetAccountsSummaryUseCase(repo)

	_, err := uc.Execute(context.Background(), "")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestGetAccountsSummary_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
		err:      repoErr,
	}
	uc := appaccount.NewGetAccountsSummaryUseCase(repo)

	_, err := uc.Execute(context.Background(), "usr_1")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
