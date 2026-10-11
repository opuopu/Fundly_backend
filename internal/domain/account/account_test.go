package account_test

import (
	"testing"

	"github.com/Dodul01/fundly/internal/domain/account"
)

func TestAccountDepositAndWithdraw(t *testing.T) {
	acc := account.New("acc-1", "usr-1", "USD")

	if err := acc.Deposit(1000); err != nil {
		t.Fatalf("unexpected error on deposit: %v", err)
	}
	if acc.Balance != 1000 {
		t.Fatalf("expected balance 1000, got %d", acc.Balance)
	}

	if err := acc.Withdraw(400); err != nil {
		t.Fatalf("unexpected error on withdraw: %v", err)
	}
	if acc.Balance != 600 {
		t.Fatalf("expected balance 600, got %d", acc.Balance)
	}

	// Insufficient balance test
	err := acc.Withdraw(1000)
	if err != account.ErrInsufficientBalance {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}
}

func TestNewAccount_Validation(t *testing.T) {
	// Success
	acc, err := account.NewAccount("acc_01", "usr_1", "Debit Card", "card", "USD", 5000, "Checking card")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if acc.Name != "Debit Card" || acc.Group != "card" || acc.Balance != 5000 {
		t.Errorf("unexpected account fields: %+v", acc)
	}

	// Empty name
	_, err = account.NewAccount("acc_02", "usr_1", "   ", "card", "USD", 0, "")
	if err != account.ErrInvalidAccountName {
		t.Errorf("expected ErrInvalidAccountName, got %v", err)
	}

	// Invalid group
	_, err = account.NewAccount("acc_03", "usr_1", "Wallet", "crypto_unknown", "USD", 0, "")
	if err != account.ErrInvalidGroup {
		t.Errorf("expected ErrInvalidGroup, got %v", err)
	}

	// Default group
	accDefault, err := account.NewAccount("acc_04", "usr_1", "Cash", "", "", 0, "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accDefault.Group != "cash" || accDefault.Currency != "USD" {
		t.Errorf("expected default group 'cash' and currency 'USD', got %+v", accDefault)
	}
}

func TestAccount_UpdateDetails(t *testing.T) {
	acc, err := account.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 0, "Initial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = acc.UpdateDetails("Primary Cash", "savings", "Updated description")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Name != "Primary Cash" || acc.Group != "savings" || acc.Description != "Updated description" {
		t.Errorf("unexpected updated fields: %+v", acc)
	}

	err = acc.UpdateDetails("", "invalid_group", "")
	if err != account.ErrInvalidGroup {
		t.Errorf("expected ErrInvalidGroup, got %v", err)
	}
}

func TestAccount_Update(t *testing.T) {
	acc, err := account.NewAccount("acc_01", "usr_1", "Cash", "cash", "USD", 0, "Initial")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	name := "Updated Name"
	group := "card"
	desc := "Updated Desc"
	err = acc.Update(&name, &group, &desc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Name != "Updated Name" || acc.Group != "card" || acc.Description != "Updated Desc" {
		t.Errorf("unexpected updated fields: %+v", acc)
	}

	// Partial update: only name
	nameOnly := "Brand New Name"
	err = acc.Update(&nameOnly, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if acc.Name != "Brand New Name" || acc.Group != "card" {
		t.Errorf("group should be unchanged: %+v", acc)
	}

	// Invalid name
	emptyName := "   "
	err = acc.Update(&emptyName, nil, nil)
	if err != account.ErrInvalidAccountName {
		t.Errorf("expected ErrInvalidAccountName, got %v", err)
	}

	// Invalid group
	invalidGroup := "crypto"
	err = acc.Update(nil, &invalidGroup, nil)
	if err != account.ErrInvalidGroup {
		t.Errorf("expected ErrInvalidGroup, got %v", err)
	}
}
