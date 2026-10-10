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
