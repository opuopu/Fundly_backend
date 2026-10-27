package budget_test

import (
	"context"
	"errors"
	"testing"

	budgetapp "github.com/Dodul01/fundly/internal/application/budget"
	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
)

func TestCreateBudget_ComputesDerivedFields(t *testing.T) {
	repo := newMockBudgetRepo()
	cats := &mockCategoryRepo{}
	uc := budgetapp.NewCreateBudgetUseCase(repo, cats, mockSpent{amount: 45480})

	res, err := uc.Execute(context.Background(), budgetapp.CreateBudgetInput{
		UserID:      "usr_1",
		Month:       "2026-09",
		LimitAmount: 80000,
		CategoryID:  "cat_food",
		Currency:    "USD",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Expenditure != 45480 {
		t.Errorf("expected expenditure 45480, got %d", res.Expenditure)
	}
	if res.RemainingAmount != 34520 {
		t.Errorf("expected remaining 34520, got %d", res.RemainingAmount)
	}
	if res.UsagePercentage != 56.85 {
		t.Errorf("expected usage 56.85, got %v", res.UsagePercentage)
	}
	if res.RemainingPercentage != 43.15 {
		t.Errorf("expected remaining percentage 43.15, got %v", res.RemainingPercentage)
	}
	if res.Category == nil || res.Category.ID != "cat_food" {
		t.Errorf("expected category ref, got %+v", res.Category)
	}
	if res.Date != "2026-09-01" {
		t.Errorf("expected date defaulted to 2026-09-01, got %s", res.Date)
	}
}

func TestCreateBudget_InvalidMonth(t *testing.T) {
	uc := budgetapp.NewCreateBudgetUseCase(newMockBudgetRepo(), &mockCategoryRepo{}, mockSpent{})
	_, err := uc.Execute(context.Background(), budgetapp.CreateBudgetInput{
		UserID:      "usr_1",
		Month:       "2026/09",
		LimitAmount: 80000,
		CategoryID:  "cat_food",
	})
	if !errors.Is(err, domainBudget.ErrInvalidMonth) {
		t.Fatalf("expected ErrInvalidMonth, got %v", err)
	}
}

func TestListBudgets_FiltersByMonth(t *testing.T) {
	repo := newMockBudgetRepo()
	b1, _ := domainBudget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")
	b2, _ := domainBudget.NewBudget("bg_2", "usr_1", "cat_food", "2026-08", 70000, "USD", "", "")
	repo.budgets["bg_1"] = b1
	repo.budgets["bg_2"] = b2

	uc := budgetapp.NewListBudgetsUseCase(repo, &mockCategoryRepo{}, mockSpent{amount: 1000})
	items, err := uc.Execute(context.Background(), "usr_1", "2026-09", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != "bg_1" {
		t.Fatalf("expected only bg_1 for 2026-09, got %+v", items)
	}
}

func TestGetBudget_OwnershipGuard(t *testing.T) {
	repo := newMockBudgetRepo()
	b, _ := domainBudget.NewBudget("bg_1", "owner", "cat_food", "2026-09", 80000, "USD", "", "")
	repo.budgets["bg_1"] = b

	uc := budgetapp.NewGetBudgetUseCase(repo, &mockCategoryRepo{}, mockSpent{})
	_, err := uc.Execute(context.Background(), "intruder", "bg_1")
	if !errors.Is(err, domainBudget.ErrBudgetNotFound) {
		t.Fatalf("expected not-found for non-owner, got %v", err)
	}
}

func TestUpdateBudget_ChangesLimit(t *testing.T) {
	repo := newMockBudgetRepo()
	b, _ := domainBudget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")
	repo.budgets["bg_1"] = b

	uc := budgetapp.NewUpdateBudgetUseCase(repo, &mockCategoryRepo{}, mockSpent{amount: 40000})
	newLimit := int64(100000)
	res, err := uc.Execute(context.Background(), budgetapp.UpdateBudgetInput{
		UserID:      "usr_1",
		ID:          "bg_1",
		LimitAmount: &newLimit,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.LimitAmount != 100000 || res.RemainingAmount != 60000 {
		t.Errorf("unexpected update result: limit=%d remaining=%d", res.LimitAmount, res.RemainingAmount)
	}
}

func TestUpdateBudget_ChangesCategoryAndMonth(t *testing.T) {
	repo := newMockBudgetRepo()
	b, _ := domainBudget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")
	repo.budgets["bg_1"] = b

	uc := budgetapp.NewUpdateBudgetUseCase(repo, &mockCategoryRepo{}, mockSpent{})
	newCategory := "cat_transport"
	newMonth := "2026-10-25"
	res, err := uc.Execute(context.Background(), budgetapp.UpdateBudgetInput{
		UserID:     "usr_1",
		ID:         "bg_1",
		CategoryID: &newCategory,
		Month:      &newMonth,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Month != "2026-10" || res.Category == nil || res.Category.ID != "cat_transport" {
		t.Errorf("unexpected update result: month=%s category=%+v", res.Month, res.Category)
	}
	if res.Date != "2026-10-25" {
		t.Errorf("expected date updated to 2026-10-25, got %s", res.Date)
	}
}

func TestDeleteBudget_OwnershipGuard(t *testing.T) {
	repo := newMockBudgetRepo()
	b, _ := domainBudget.NewBudget("bg_1", "owner", "cat_food", "2026-09", 80000, "USD", "", "")
	repo.budgets["bg_1"] = b

	uc := budgetapp.NewDeleteBudgetUseCase(repo)
	if err := uc.Execute(context.Background(), "intruder", "bg_1"); !errors.Is(err, domainBudget.ErrBudgetNotFound) {
		t.Fatalf("expected not-found for non-owner, got %v", err)
	}
	if err := uc.Execute(context.Background(), "owner", "bg_1"); err != nil {
		t.Fatalf("owner delete failed: %v", err)
	}
}
