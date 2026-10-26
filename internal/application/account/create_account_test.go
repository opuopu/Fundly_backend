package account_test

import (
	"context"
	"errors"
	"testing"

	appaccount "github.com/Dodul01/fundly/internal/application/account"
	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

func TestCreateAccount_Success(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	
	uc := appaccount.NewCreateAccountUseCase(repo)

	input := appaccount.CreateAccountInput{
		UserID:         "usr_alice",
		Name:           "Cash Wallet",
		Group:          "cash",
		InitialBalance: 20000,
		Currency:       "USD",
		Description:    "Physical wallet cash",
	}

	res, err := uc.Execute(context.Background(), input)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID == "" {
		t.Errorf("expected non-empty account ID")
	}
	
	if res.Name != "Cash Wallet" {
		t.Errorf("expected name 'Cash Wallet', got '%s'", res.Name)
	}
	
	if res.Group != "cash" {
		t.Errorf("expected group 'cash', got '%s'", res.Group)
	}
	
	if res.Balance != 20000 {
		t.Errorf("expected balance 20000, got %d", res.Balance)
	}
	
	if res.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", res.Currency)
	}
	
	if res.Description != "Physical wallet cash" {
		t.Errorf("expected description 'Physical wallet cash', got '%s'", res.Description)
	}

	// Verify persistence in repo
	saved, ok := repo.accounts[res.ID]
	
	if !ok {
		t.Fatalf("expected account to be saved in repository")
	}
	
	if saved.Balance != 20000 || saved.Name != "Cash Wallet" {
		t.Errorf("saved account attributes do not match")
	}
}

func TestCreateAccount_Defaults(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	
	uc := appaccount.NewCreateAccountUseCase(repo)

	input := appaccount.CreateAccountInput{
		UserID:         "usr_alice",
		Name:           "My Savings",
		Group:          "", // should default to "cash"
		InitialBalance: 0,
		Currency:       "", // should default to "USD"
	}

	res, err := uc.Execute(context.Background(), input)
	
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Group != "cash" {
		t.Errorf("expected default group 'cash', got '%s'", res.Group)
	}
	if res.Currency != "USD" {
		t.Errorf("expected default currency 'USD', got '%s'", res.Currency)
	}
	if res.Balance != 0 {
		t.Errorf("expected balance 0, got %d", res.Balance)
	}
}

func TestCreateAccount_InvalidName(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	
	uc := appaccount.NewCreateAccountUseCase(repo)

	input := appaccount.CreateAccountInput{
		UserID: "usr_alice",
		Name:   "   ",
		Group:  "cash",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainAccount.ErrInvalidAccountName) {
		t.Fatalf("expected ErrInvalidAccountName, got %v", err)
	}
}

func TestCreateAccount_InvalidGroup(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewCreateAccountUseCase(repo)

	input := appaccount.CreateAccountInput{
		UserID: "usr_alice",
		Name:   "Crypto Wallet",
		Group:  "crypto",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainAccount.ErrInvalidGroup) {
		t.Fatalf("expected ErrInvalidGroup, got %v", err)
	}
}

func TestCreateAccount_MissingUserID(t *testing.T) {
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
	}
	uc := appaccount.NewCreateAccountUseCase(repo)

	input := appaccount.CreateAccountInput{
		UserID: "",
		Name:   "Emergency Fund",
		Group:  "savings",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainAccount.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}

func TestCreateAccount_RepoError(t *testing.T) {
	repoErr := errors.New("db insert failure")
	repo := &mockAccountRepo{
		accounts: make(map[string]*domainAccount.Account),
		err:      repoErr,
	}
	uc := appaccount.NewCreateAccountUseCase(repo)

	input := appaccount.CreateAccountInput{
		UserID: "usr_alice",
		Name:   "Checking Account",
		Group:  "checking",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
