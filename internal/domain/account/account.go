package account

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
	ErrInvalidAccountName  = errors.New("account name cannot be empty")
	ErrInvalidGroup        = errors.New("invalid account group")
	ErrNoFieldsToUpdate    = errors.New("no fields to update")
)

var ValidGroups = map[string]bool{
	"cash":       true,
	"card":       true,
	"savings":    true,
	"investment": true,
	"loan":       true,
	"insurance":  true,
	"checking":   true,
}

type Account struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Group       string    `json:"group"`
	Currency    string    `json:"currency"`
	Balance     int64     `json:"balance"` // balance stored in minor units (e.g. cents) to avoid float rounding errors
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func New(id, userID, currency string) *Account {
	now := time.Now().UTC()
	if currency == "" {
		currency = "USD"
	}
	return &Account{
		ID:        id,
		UserID:    userID,
		Name:      "Cash",
		Group:     "cash",
		Currency:  currency,
		Balance:   0,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func NewAccount(id, userID, name, group, currency string, initialBalance int64, description string) (*Account, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrInvalidAccountName
	}

	group = strings.ToLower(strings.TrimSpace(group))
	if group == "" {
		group = "cash"
	} else if !ValidGroups[group] {
		return nil, ErrInvalidGroup
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}

	now := time.Now().UTC()
	return &Account{
		ID:          id,
		UserID:      userID,
		Name:        name,
		Group:       group,
		Currency:    currency,
		Balance:     initialBalance,
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (a *Account) UpdateDetails(name, group, description string) error {
	name = strings.TrimSpace(name)
	if name != "" {
		a.Name = name
	}
	group = strings.ToLower(strings.TrimSpace(group))
	if group != "" {
		if !ValidGroups[group] {
			return ErrInvalidGroup
		}
		a.Group = group
	}
	if description != "" {
		a.Description = strings.TrimSpace(description)
	}
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Update(name, group, description *string) error {
	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return ErrInvalidAccountName
		}
		a.Name = trimmedName
	}
	if group != nil {
		trimmedGroup := strings.ToLower(strings.TrimSpace(*group))
		if !ValidGroups[trimmedGroup] {
			return ErrInvalidGroup
		}
		a.Group = trimmedGroup
	}
	if description != nil {
		a.Description = strings.TrimSpace(*description)
	}
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	a.Balance += amount
	a.UpdatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Balance < amount {
		return ErrInsufficientBalance
	}
	a.Balance -= amount
	a.UpdatedAt = time.Now().UTC()
	return nil
}
