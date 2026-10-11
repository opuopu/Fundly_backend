package account_test

import (
	"context"
	"errors"
	"testing"

	appaccount "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

func TestGetAccount_Success(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Checking Account", "checking", "USD", 150000, "Primary checking")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewGetAccountUseCase(repo)

	res, err := uc.Execute(context.Background(), "usr_1", "acc_01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != "acc_01" {
		t.Errorf("expected ID 'acc_01', got '%s'", res.ID)
	}
	if res.Name != "Checking Account" {
		t.Errorf("expected name 'Checking Account', got '%s'", res.Name)
	}
	if res.Group != "checking" {
		t.Errorf("expected group 'checking', got '%s'", res.Group)
	}
	if res.Balance != 150000 {
		t.Errorf("expected balance 150000, got %d", res.Balance)
	}
	if res.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", res.Currency)
	}
	if res.Description != "Primary checking" {
		t.Errorf("expected description 'Primary checking', got '%s'", res.Description)
	}
}

func TestGetAccount_NotFound(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewGetAccountUseCase(repo)

	_, err := uc.Execute(context.Background(), "usr_1", "non_existent")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestGetAccount_OtherUserAccount(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	// Belongs to usr_2
	acc, _ := domainAccount.NewAccount("acc_02", "usr_2", "Secret Account", "savings", "USD", 1000000, "")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewGetAccountUseCase(repo)

	// Requested by usr_1
	_, err := uc.Execute(context.Background(), "usr_1", "acc_02")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound for unauthorized access, got %v", err)
	}
}

func TestGetAccount_EmptyIDs(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewGetAccountUseCase(repo)

	_, err := uc.Execute(context.Background(), "", "acc_01")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Errorf("expected ErrAccountNotFound on empty userID, got %v", err)
	}

	_, err = uc.Execute(context.Background(), "usr_1", "")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Errorf("expected ErrAccountNotFound on empty accountID, got %v", err)
	}
}

func TestGetAccount_RepoError(t *testing.T) {
	repoErr := errors.New("db query error")
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
		err:      repoErr,
	}
	uc := appaccount.NewGetAccountUseCase(repo)

	_, err := uc.Execute(context.Background(), "usr_1", "acc_01")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
