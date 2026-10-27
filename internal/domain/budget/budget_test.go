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
	if got := b.Date.Format("2006-01-02"); got != "2026-09-01" {
		t.Errorf("expected date defaulted to 2026-09-01, got %s", got)
	}
}

func TestNewBudget_AcceptsFullDateForMonth(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		wantDate string
	}{
		{"full date", "2026-09-16", "2026-09-16"},
		{"iso timestamp", "2026-09-16T00:00:00.000Z", "2026-09-16"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := budget.NewBudget("bg_1", "usr_1", "cat_food", tc.input, 80000, "USD", "", "")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if b.Month != "2026-09" {
				t.Errorf("expected month normalized to 2026-09, got %s", b.Month)
			}
			if got := b.Date.Format("2006-01-02"); got != tc.wantDate {
				t.Errorf("expected date %s, got %s", tc.wantDate, got)
			}
		})
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

	if err := b.Update(nil, nil, nil, nil, nil, nil); !errors.Is(err, budget.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}

	newLimit := int64(90000)
	if err := b.Update(nil, nil, &newLimit, nil, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.LimitAmount != 90000 {
		t.Errorf("expected limit 90000, got %d", b.LimitAmount)
	}

	zero := int64(0)
	if err := b.Update(nil, nil, &zero, nil, nil, nil); !errors.Is(err, budget.ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit, got %v", err)
	}
}

func TestBudget_Update_CategoryAndMonth(t *testing.T) {
	b, _ := budget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")

	newCategory := "cat_transport"
	newMonth := "2026-10-25"
	if err := b.Update(&newCategory, &newMonth, nil, nil, nil, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.CategoryID != "cat_transport" || b.Month != "2026-10" {
		t.Errorf("expected category/month updated, got %s/%s", b.CategoryID, b.Month)
	}
	if got := b.Date.Format("2006-01-02"); got != "2026-10-25" {
		t.Errorf("expected date updated to 2026-10-25, got %s", got)
	}

	badMonth := "2026/10"
	if err := b.Update(nil, &badMonth, nil, nil, nil, nil); !errors.Is(err, budget.ErrInvalidMonth) {
		t.Fatalf("expected ErrInvalidMonth, got %v", err)
	}

	emptyCategory := "  "
	if err := b.Update(&emptyCategory, nil, nil, nil, nil, nil); !errors.Is(err, budget.ErrCategoryRequired) {
		t.Fatalf("expected ErrCategoryRequired, got %v", err)
	}
}

func TestBudget_Update_Expenditure(t *testing.T) {
	b, _ := budget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")
	if b.ManualExpenditure != nil {
		t.Fatalf("expected no manual expenditure on creation, got %v", *b.ManualExpenditure)
	}

	manual := int64(3000)
	if err := b.Update(nil, nil, nil, nil, nil, &manual); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ManualExpenditure == nil || *b.ManualExpenditure != 3000 {
		t.Errorf("expected manual expenditure 3000, got %v", b.ManualExpenditure)
	}

	negative := int64(-1)
	if err := b.Update(nil, nil, nil, nil, nil, &negative); !errors.Is(err, budget.ErrInvalidExpenditure) {
		t.Fatalf("expected ErrInvalidExpenditure, got %v", err)
	}
}
