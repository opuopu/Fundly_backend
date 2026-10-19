package stat_test

import (
	"context"
	"errors"
	"testing"

	statapp "github.com/Dodul01/fundly/internal/application/stat"
	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	"github.com/Dodul01/fundly/internal/domain/category"
	domainConfig "github.com/Dodul01/fundly/internal/domain/config"
	domainStat "github.com/Dodul01/fundly/internal/domain/stat"
)

func TestBreakdown_ComputesPercentages(t *testing.T) {
	stats := &mockStatRepo{breakdown: []domainStat.CategoryAmount{
		{CategoryID: "cat_salary", Amount: 5600},
		{CategoryID: "cat_kpi", Amount: 5600},
		{CategoryID: "cat_other", Amount: 8800},
	}}
	cats := &mockCategoryRepo{categories: map[string]*category.Category{
		"cat_salary": {ID: "cat_salary", Name: "Salary", Color: "#4CAF50"},
	}}

	uc := statapp.NewBreakdownUseCase(stats, cats, nil)
	res, err := uc.Execute(context.Background(), statapp.BreakdownInput{
		UserID: "usr_1",
		Type:   "income",
		Month:  "2026-08",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.TotalAmount != 20000 {
		t.Errorf("expected total 20000, got %d", res.TotalAmount)
	}
	if res.Type != "INCOME" {
		t.Errorf("expected type INCOME, got %s", res.Type)
	}
	if len(res.Breakdown) != 3 {
		t.Fatalf("expected 3 slices, got %d", len(res.Breakdown))
	}
	if res.Breakdown[0].Percentage != 28.0 {
		t.Errorf("expected 28.0%%, got %v", res.Breakdown[0].Percentage)
	}
	if res.Breakdown[0].CategoryName != "Salary" || res.Breakdown[0].Color != "#4CAF50" {
		t.Errorf("expected enriched category, got %+v", res.Breakdown[0])
	}
}

func TestBreakdown_InvalidType(t *testing.T) {
	uc := statapp.NewBreakdownUseCase(&mockStatRepo{}, &mockCategoryRepo{}, nil)
	_, err := uc.Execute(context.Background(), statapp.BreakdownInput{UserID: "usr_1", Type: "transfer"})
	if !errors.Is(err, statapp.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestBreakdown_DefaultsMonth(t *testing.T) {
	uc := statapp.NewBreakdownUseCase(&mockStatRepo{}, &mockCategoryRepo{}, nil)
	res, err := uc.Execute(context.Background(), statapp.BreakdownInput{UserID: "usr_1", Type: "expense"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Month != statapp.CurrentMonth() {
		t.Errorf("expected default month %s, got %s", statapp.CurrentMonth(), res.Month)
	}
}

type stubConfigReader struct{ currency string }

func (s stubConfigReader) Get(_ context.Context, userID string) (*domainConfig.Config, error) {
	return &domainConfig.Config{UserID: userID, DefaultCurrency: s.currency}, nil
}

func TestBreakdown_ResolvesUserCurrency(t *testing.T) {
	stats := &mockStatRepo{breakdown: []domainStat.CategoryAmount{{CategoryID: "cat_food", Amount: 100}}}
	uc := statapp.NewBreakdownUseCase(stats, &mockCategoryRepo{}, stubConfigReader{currency: "EUR"})

	res, err := uc.Execute(context.Background(), statapp.BreakdownInput{UserID: "usr_1", Type: "expense", Month: "2026-08"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Currency != "EUR" {
		t.Errorf("expected EUR from user config, got %s", res.Currency)
	}
}

func TestOverview_TotalsAndUtilization(t *testing.T) {
	stats := &mockStatRepo{totals: domainStat.Totals{TotalIncome: 100000, TotalExpense: 40000}}
	b, _ := domainBudget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")
	budgets := &mockBudgetRepo{budgets: []*domainBudget.Budget{b}}

	uc := statapp.NewOverviewUseCase(stats, budgets, nil)
	res, err := uc.Execute(context.Background(), "usr_1", "2026-09")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.NetBalance != 60000 {
		t.Errorf("expected net 60000, got %d", res.NetBalance)
	}
	if res.BudgetLimit != 80000 {
		t.Errorf("expected budget limit 80000, got %d", res.BudgetLimit)
	}
	if res.BudgetUtilization != 50.0 {
		t.Errorf("expected utilization 50.0, got %v", res.BudgetUtilization)
	}
}
