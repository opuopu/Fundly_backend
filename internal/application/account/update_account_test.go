package account_test

import (
	"context"
	"errors"
	"testing"

	appaccount "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

func TestUpdateAccount_Success(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewUpdateAccountUseCase(repo)

	newName := "Primary Cash Wallet"
	newGroup := "card"
	newDesc := "Updated description"
	input := appaccount.UpdateAccountInput{
		UserID:      "usr_1",
		AccountID:   "acc_01",
		Name:        &newName,
		Group:       &newGroup,
		Description: &newDesc,
	}

	res, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Primary Cash Wallet" {
		t.Errorf("expected name 'Primary Cash Wallet', got '%s'", res.Name)
	}
	if res.Group != "card" {
		t.Errorf("expected group 'card', got '%s'", res.Group)
	}
	if res.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got '%s'", res.Description)
	}

	// Verify persistence in repo
	saved := repo.accounts["acc_01"]
	if saved.Name != "Primary Cash Wallet" || saved.Group != "card" {
		t.Errorf("repo account was not updated correctly: %+v", saved)
	}
}

func TestUpdateAccount_PartialUpdate(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewUpdateAccountUseCase(repo)

	newName := "Renamed Cash"
	input := appaccount.UpdateAccountInput{
		UserID:    "usr_1",
		AccountID: "acc_01",
		Name:      &newName,
	}

	res, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Renamed Cash" {
		t.Errorf("expected name 'Renamed Cash', got '%s'", res.Name)
	}
	if res.Group != "cash" {
		t.Errorf("group should remain 'cash', got '%s'", res.Group)
	}
	if res.Description != "Old description" {
		t.Errorf("description should remain 'Old description', got '%s'", res.Description)
	}
}

func TestUpdateAccount_NoFieldsToUpdate(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewUpdateAccountUseCase(repo)

	input := appaccount.UpdateAccountInput{
		UserID:    "usr_1",
		AccountID: "acc_01",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainAccount.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}
}

func TestUpdateAccount_InvalidName(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewUpdateAccountUseCase(repo)

	emptyName := "   "
	input := appaccount.UpdateAccountInput{
		UserID:    "usr_1",
		AccountID: "acc_01",
		Name:      &emptyName,
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainAccount.ErrInvalidAccountName) {
		t.Fatalf("expected ErrInvalidAccountName, got %v", err)
	}
}

func TestUpdateAccount_InvalidGroup(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 20000, "Old description")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewUpdateAccountUseCase(repo)

	invalidGroup := "crypto"
	input := appaccount.UpdateAccountInput{
		UserID:    "usr_1",
		AccountID: "acc_01",
		Group:     &invalidGroup,
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainAccount.ErrInvalidGroup) {
		t.Fatalf("expected ErrInvalidGroup, got %v", err)
	}
}

func TestUpdateAccount_NotFoundOrOtherUser(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	acc, _ := domainAccount.NewAccount("acc_02", "usr_2", "Secret", "savings", "USD", 5000, "")
	repo.accounts[acc.ID] = acc

	uc := appaccount.NewUpdateAccountUseCase(repo)

	name := "Hacked Name"
	// usr_1 tries to update usr_2's account
	input := appaccount.UpdateAccountInput{
		UserID:    "usr_1",
		AccountID: "acc_02",
		Name:      &name,
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}
