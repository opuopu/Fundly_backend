package transaction_test

import (
	"errors"
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/domain/transaction"
)

func TestNewEntry_Income(t *testing.T) {
	tx, err := transaction.NewEntry("tx_1", "usr_1", "income", time.Time{}, 5000, "usd", "cat_salary", "acc_1", "pay", "monthly")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.Type != transaction.TypeIncome {
		t.Errorf("expected INCOME, got %s", tx.Type)
	}
	if tx.ToAccountID != "acc_1" || tx.FromAccountID != "" {
		t.Errorf("income should credit ToAccountID, got from=%q to=%q", tx.FromAccountID, tx.ToAccountID)
	}
	if tx.Currency != "USD" {
		t.Errorf("expected currency normalized to USD, got %s", tx.Currency)
	}
	if tx.Date.IsZero() {
		t.Error("expected date defaulted to now")
	}
}

func TestNewEntry_Expense(t *testing.T) {
	tx, err := transaction.NewEntry("tx_1", "usr_1", "EXPENSE", time.Now(), 2000, "USD", "cat_food", "acc_1", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tx.FromAccountID != "acc_1" || tx.ToAccountID != "" {
		t.Errorf("expense should debit FromAccountID, got from=%q to=%q", tx.FromAccountID, tx.ToAccountID)
	}
}

func TestNewEntry_Validation(t *testing.T) {
	cases := []struct {
		name                                    string
		txType, currency, categoryID, accountID string
		amount                                  int64
		want                                    error
	}{
		{"bad type", "transfer", "USD", "cat_food", "acc_1", 100, transaction.ErrInvalidType},
		{"zero amount", "expense", "USD", "cat_food", "acc_1", 0, transaction.ErrInvalidAmount},
		{"missing account", "expense", "USD", "cat_food", "", 100, transaction.ErrAccountRequired},
		{"missing category", "expense", "USD", "", "acc_1", 100, transaction.ErrCategoryRequired},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := transaction.NewEntry("tx_1", "usr_1", tc.txType, time.Now(), tc.amount, tc.currency, tc.categoryID, tc.accountID, "", "")
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestUpdate_NoFields(t *testing.T) {
	tx, _ := transaction.NewEntry("tx_1", "usr_1", "expense", time.Now(), 2000, "USD", "cat_food", "acc_1", "", "")
	if err := tx.Update(nil, nil, nil, nil, nil, nil); !errors.Is(err, transaction.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}
}
