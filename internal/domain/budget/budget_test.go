package budget_test

import (
	"errors"
	"testing"

	"github.com/Dodul01/fundly/internal/domain/budget"
)

func TestNewBudget_Success(t *testing.T) {
	b, err := budget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "usd", "food", "monthly food")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.Currency != "USD" {
		t.Errorf("expected currency normalized to USD, got %s", b.Currency)
	}
	if b.Month != "2026-09" {
		t.Errorf("unexpected month %s", b.Month)
	}
}

func TestNewBudget_Validation(t *testing.T) {
	cases := []struct {
		name              string
		categoryID, month string
		limit             int64
		want              error
	}{
		{"missing category", "", "2026-09", 100, budget.ErrCategoryRequired},
		{"bad month", "cat_food", "2026/09", 100, budget.ErrInvalidMonth},
		{"month 13", "cat_food", "2026-13", 100, budget.ErrInvalidMonth},
		{"zero limit", "cat_food", "2026-09", 0, budget.ErrInvalidLimit},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := budget.NewBudget("bg_1", "usr_1", tc.categoryID, tc.month, tc.limit, "USD", "", "")
			if !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestBudget_Update(t *testing.T) {
	b, _ := budget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")

	if err := b.Update(nil, nil, nil); !errors.Is(err, budget.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}

	newLimit := int64(90000)
	if err := b.Update(&newLimit, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.LimitAmount != 90000 {
		t.Errorf("expected limit 90000, got %d", b.LimitAmount)
	}

	zero := int64(0)
	if err := b.Update(&zero, nil, nil); !errors.Is(err, budget.ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
}
