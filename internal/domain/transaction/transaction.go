package transaction

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInvalidType         = errors.New("invalid transaction type: must be income or expense")
	ErrAccountRequired     = errors.New("account_id is required")
	ErrCategoryRequired    = errors.New("category_id is required")
	ErrNoFieldsToUpdate    = errors.New("at least one field must be provided to update")
)

type Type string

const (
	TypeDeposit  Type = "DEPOSIT"
	TypeWithdraw Type = "WITHDRAW"
	TypeTransfer Type = "TRANSFER"
	TypeIncome   Type = "INCOME"
	TypeExpense  Type = "EXPENSE"
)

// ParseType normalizes and validates a user-provided income/expense type.
func ParseType(s string) (Type, error) {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "INCOME":
		return TypeIncome, nil
	case "EXPENSE":
		return TypeExpense, nil
	default:
		return "", ErrInvalidType
	}
}

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusCompleted Status = "COMPLETED"
	StatusFailed    Status = "FAILED"
)

type Transaction struct {
	ID            string    `json:"id"`
	UserID        string    `json:"user_id,omitempty"`
	FromAccountID string    `json:"from_account_id,omitempty"`
	ToAccountID   string    `json:"to_account_id,omitempty"`
	CategoryID    string    `json:"category_id,omitempty"`
	Amount        int64     `json:"amount"`
	Currency      string    `json:"currency"`
	Type          Type      `json:"type"`
	Status        Status    `json:"status"`
	Note          string    `json:"note,omitempty"`
	Description   string    `json:"description,omitempty"`
	Date          time.Time `json:"date"`
	CreatedAt     time.Time `json:"created_at"`
}

// NewEntry builds a validated income/expense transaction. For INCOME the amount is
// credited to the referenced account (ToAccountID); for EXPENSE it is debited from the
// account (FromAccountID), which keeps the account-scoped queries and balance effects
// consistent with the transfer flow.
func NewEntry(id, userID, txType string, date time.Time, amount int64, currency, categoryID, accountID, note, description string) (*Transaction, error) {
	t, err := ParseType(txType)
	if err != nil {
		return nil, err
	}

	if amount <= 0 {
		return nil, ErrInvalidAmount
	}

	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return nil, ErrAccountRequired
	}

	categoryID = strings.TrimSpace(categoryID)
	if categoryID == "" {
		return nil, ErrCategoryRequired
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}

	if date.IsZero() {
		date = time.Now().UTC()
	}

	tx := &Transaction{
		ID:          id,
		UserID:      userID,
		CategoryID:  categoryID,
		Amount:      amount,
		Currency:    currency,
		Type:        t,
		Status:      StatusCompleted,
		Note:        strings.TrimSpace(note),
		Description: strings.TrimSpace(description),
		Date:        date.UTC(),
		CreatedAt:   time.Now().UTC(),
	}

	if t == TypeIncome {
		tx.ToAccountID = accountID
	} else {
		tx.FromAccountID = accountID
	}

	return tx, nil
}

// AccountID returns the account this entry affects regardless of income/expense direction.
func (t *Transaction) AccountID() string {
	if t.ToAccountID != "" {
		return t.ToAccountID
	}
	return t.FromAccountID
}

// Update applies the provided non-nil fields to an existing entry.
func (t *Transaction) Update(amount *int64, date *time.Time, categoryID, accountID, note, description *string) error {
	if amount == nil && date == nil && categoryID == nil && accountID == nil && note == nil && description == nil {
		return ErrNoFieldsToUpdate
	}

	if amount != nil {
		if *amount <= 0 {
			return ErrInvalidAmount
		}
		t.Amount = *amount
	}
	if date != nil && !date.IsZero() {
		t.Date = date.UTC()
	}
	if categoryID != nil {
		trimmed := strings.TrimSpace(*categoryID)
		if trimmed == "" {
			return ErrCategoryRequired
		}
		t.CategoryID = trimmed
	}
	if accountID != nil {
		trimmed := strings.TrimSpace(*accountID)
		if trimmed == "" {
			return ErrAccountRequired
		}
		if t.Type == TypeIncome {
			t.ToAccountID = trimmed
		} else {
			t.FromAccountID = trimmed
		}
	}
	if note != nil {
		t.Note = strings.TrimSpace(*note)
	}
	if description != nil {
		t.Description = strings.TrimSpace(*description)
	}

	return nil
}
