package account_test

import (
	"context"
	"errors"
	"testing"

	appaccount "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

func TestDeleteAccount_Success(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewDeleteAccountUseCase(repo)

	err := uc.Execute(context.Background(), "usr_1", "acc_01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, exists := repo.accounts["acc_01"]; exists {
		t.Errorf("expected account to be deleted from repo")
	}
}

func TestDeleteAccount_NotFound(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewDeleteAccountUseCase(repo)

	err := uc.Execute(context.Background(), "usr_1", "non_existent")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestDeleteAccount_OtherUserAccount(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_02", "usr_2", "Secret", "savings", "USD", 5000, "")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewDeleteAccountUseCase(repo)

	// usr_1 tries to delete usr_2's account
	err := uc.Execute(context.Background(), "usr_1", "acc_02")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}

	// Verify account still exists for usr_2
	if _, exists := repo.accounts["acc_02"]; !exists {
		t.Errorf("usr_2's account should not have been deleted")
	}
}

func TestDeleteAccount_EmptyIDs(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewDeleteAccountUseCase(repo)

	err := uc.Execute(context.Background(), "", "acc_01")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Errorf("expected ErrAccountNotFound on empty userID, got %v", err)
	}

	err = uc.Execute(context.Background(), "usr_1", "")
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Errorf("expected ErrAccountNotFound on empty accountID, got %v", err)
	}
}

func TestDeleteAccount_RepoError(t *testing.T) {
	repoErr := errors.New("db error")
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
		err:      repoErr,
	}
	uc := appaccount.NewDeleteAccountUseCase(repo)

	err := uc.Execute(context.Background(), "usr_1", "acc_01")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
