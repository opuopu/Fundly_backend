package budget

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrBudgetNotFound   = errors.New("budget not found")
	ErrInvalidMonth     = errors.New("month must be in YYYY-MM format")
	ErrInvalidLimit     = errors.New("limit amount must be greater than zero")
	ErrCategoryRequired = errors.New("category_id is required")
	ErrNoFieldsToUpdate = errors.New("at least one field must be provided to update")
	ErrBudgetExists     = errors.New("a budget already exists for this category and month")
)

var monthPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// ValidMonth reports whether s is a well-formed YYYY-MM value.
func ValidMonth(s string) bool {
	return monthPattern.MatchString(strings.TrimSpace(s))
}

type Budget struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CategoryID  string    `json:"category_id"`
	Month       string    `json:"month"`
	LimitAmount int64     `json:"limit_amount"`
	Currency    string    `json:"currency"`
	Note        string    `json:"note,omitempty"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewBudget(id, userID, categoryID, month string, limitAmount int64, currency, note, description string) (*Budget, error) {
	categoryID = strings.TrimSpace(categoryID)
	if categoryID == "" {
		return nil, ErrCategoryRequired
	}

	month = strings.TrimSpace(month)
	if !ValidMonth(month) {
		return nil, ErrInvalidMonth
	}

	if limitAmount <= 0 {
		return nil, ErrInvalidLimit
	}

	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "" {
		currency = "USD"
	}

	now := time.Now().UTC()
	return &Budget{
		ID:          id,
		UserID:      userID,
		CategoryID:  categoryID,
		Month:       month,
		LimitAmount: limitAmount,
		Currency:    currency,
		Note:        strings.TrimSpace(note),
		Description: strings.TrimSpace(description),
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// Update applies the provided non-nil fields. categoryID and month together identify the
// allocation, so changing either is allowed but re-validated the same way NewBudget does,
// since the result still has to be a well-formed, uniquely-keyed allocation.
func (b *Budget) Update(categoryID, month *string, limitAmount *int64, note, description *string) error {
	if categoryID == nil && month == nil && limitAmount == nil && note == nil && description == nil {
		return ErrNoFieldsToUpdate
	}
	if categoryID != nil {
		trimmed := strings.TrimSpace(*categoryID)
		if trimmed == "" {
			return ErrCategoryRequired
		}
		b.CategoryID = trimmed
	}
	if month != nil {
		trimmed := strings.TrimSpace(*month)
		if !ValidMonth(trimmed) {
			return ErrInvalidMonth
		}
		b.Month = trimmed
	}
	if limitAmount != nil {
		if *limitAmount <= 0 {
			return ErrInvalidLimit
		}
		b.LimitAmount = *limitAmount
	}
	if note != nil {
		b.Note = strings.TrimSpace(*note)
	}
	if description != nil {
		b.Description = strings.TrimSpace(*description)
	}
	b.UpdatedAt = time.Now().UTC()
	return nil
}
