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

// NormalizeMonth accepts either YYYY-MM or a full date/timestamp (YYYY-MM-DD, optionally with
// a time component, as date pickers commonly submit) and returns its YYYY-MM prefix. Inputs
// that don't start with a recognizable YYYY-MM are returned unchanged, so ValidMonth still
// rejects them with the usual error.
func NormalizeMonth(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 7 && ValidMonth(s[:7]) {
		return s[:7]
	}
	return s
}

var dateLayouts = []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"}

// parseExactDate tries to parse s as a full calendar date (optionally with a time/zone
// component). It returns false when s doesn't carry day-level precision (e.g. plain "2026-10").
func parseExactDate(s string) (time.Time, bool) {
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// resolveMonthAndDate interprets the client-supplied value (YYYY-MM, YYYY-MM-DD, or an ISO
// timestamp) and returns both the YYYY-MM identity used for budget uniqueness/filtering and the
// specific calendar date to display, defaulting to the 1st when only a month was given.
func resolveMonthAndDate(raw string) (month string, date time.Time, err error) {
	month = NormalizeMonth(raw)
	if !ValidMonth(month) {
		return "", time.Time{}, ErrInvalidMonth
	}
	if d, ok := parseExactDate(strings.TrimSpace(raw)); ok {
		return month, d, nil
	}
	firstOfMonth, _ := time.Parse("2006-01", month)
	return month, firstOfMonth, nil
}

type Budget struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	CategoryID  string    `json:"category_id"`
	Month       string    `json:"month"`
	Date        time.Time `json:"date"`
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

	resolvedMonth, date, err := resolveMonthAndDate(month)
	if err != nil {
		return nil, err
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
		Month:       resolvedMonth,
		Date:        date,
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
		resolvedMonth, date, err := resolveMonthAndDate(*month)
		if err != nil {
			return err
		}
		b.Month = resolvedMonth
		b.Date = date
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
