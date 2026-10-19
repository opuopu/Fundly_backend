package report_test

import (
	"context"
	"errors"
	"testing"

	reportapp "github.com/Dodul01/fundly/internal/application/report"
	"github.com/Dodul01/fundly/internal/domain/category"
	domainStat "github.com/Dodul01/fundly/internal/domain/stat"
)

// mockStatRepo returns per-month breakdowns keyed by month.
type mockStatRepo struct {
	byMonth map[string][]domainStat.CategoryAmount
}

func (m *mockStatRepo) BreakdownByCategory(_ context.Context, _, _, month, _ string) ([]domainStat.CategoryAmount, error) {
	return m.byMonth[month], nil
}
func (m *mockStatRepo) MonthlyTotals(_ context.Context, _, _ string) (domainStat.Totals, error) {
	return domainStat.Totals{}, nil
}

type mockCategoryRepo struct {
	categories map[string]*category.Category
}

func (m *mockCategoryRepo) GetAll(_ context.Context) ([]*category.Category, error) { return nil, nil }
func (m *mockCategoryRepo) GetByType(_ context.Context, _ category.CategoryType) ([]*category.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) GetByID(_ context.Context, id string) (*category.Category, error) {
	if c, ok := m.categories[id]; ok {
		return c, nil
	}
	return nil, category.ErrCategoryNotFound
}
func (m *mockCategoryRepo) Create(_ context.Context, _ *category.Category) error { return nil }
func (m *mockCategoryRepo) Update(_ context.Context, _ *category.Category) error { return nil }
func (m *mockCategoryRepo) Delete(_ context.Context, _ string) error            { return nil }

func TestCompare_SelectedCategories(t *testing.T) {
	stats := &mockStatRepo{byMonth: map[string][]domainStat.CategoryAmount{
		"2026-08": {{CategoryID: "cat_salary", Amount: 24300}, {CategoryID: "cat_allowance", Amount: 24300}},
		"2026-07": {{CategoryID: "cat_salary", Amount: 20000}, {CategoryID: "cat_allowance", Amount: 24300}},
	}}
	cats := &mockCategoryRepo{categories: map[string]*category.Category{
		"cat_salary": {ID: "cat_salary", Name: "Salary"},
	}}

	uc := reportapp.NewCompareUseCase(stats, cats, nil)
	res, err := uc.Execute(context.Background(), reportapp.CompareInput{
		UserID:          "usr_1",
		Type:            "income",
		PrimaryMonth:    "2026-08",
		ComparisonMonth: "2026-07",
		CategoryIDs:     []string{"cat_salary", "cat_allowance"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.PrimaryTotal != 48600 || res.ComparisonTotal != 44300 {
		t.Errorf("unexpected totals: primary=%d comparison=%d", res.PrimaryTotal, res.ComparisonTotal)
	}
	if res.ComparedBalance != 4300 {
		t.Errorf("expected compared_balance 4300, got %d", res.ComparedBalance)
	}
	if len(res.Categories) != 2 {
		t.Fatalf("expected 2 category rows, got %d", len(res.Categories))
	}
	// First selected category keeps its order and computes difference.
	if res.Categories[0].CategoryID != "cat_salary" || res.Categories[0].Difference != 4300 {
		t.Errorf("unexpected first row: %+v", res.Categories[0])
	}
	if res.Categories[0].Name != "Salary" {
		t.Errorf("expected enriched name Salary, got %q", res.Categories[0].Name)
	}
}

func TestCompare_UnionWhenNoCategoryFilter(t *testing.T) {
	stats := &mockStatRepo{byMonth: map[string][]domainStat.CategoryAmount{
		"2026-08": {{CategoryID: "cat_food", Amount: 5000}},
		"2026-07": {{CategoryID: "cat_shopping", Amount: 3000}},
	}}
	uc := reportapp.NewCompareUseCase(stats, &mockCategoryRepo{}, nil)
	res, err := uc.Execute(context.Background(), reportapp.CompareInput{
		UserID:          "usr_1",
		Type:            "expense",
		PrimaryMonth:    "2026-08",
		ComparisonMonth: "2026-07",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Categories) != 2 {
		t.Fatalf("expected union of 2 categories, got %d", len(res.Categories))
	}
}

func TestCompare_InvalidType(t *testing.T) {
	uc := reportapp.NewCompareUseCase(&mockStatRepo{}, &mockCategoryRepo{}, nil)
	_, err := uc.Execute(context.Background(), reportapp.CompareInput{UserID: "usr_1", Type: "bogus"})
	if !errors.Is(err, reportapp.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}

func TestCompare_InvalidMonth(t *testing.T) {
	uc := reportapp.NewCompareUseCase(&mockStatRepo{}, &mockCategoryRepo{}, nil)
	_, err := uc.Execute(context.Background(), reportapp.CompareInput{
		UserID:       "usr_1",
		Type:         "income",
		PrimaryMonth: "2026/08",
	})
	if !errors.Is(err, reportapp.ErrInvalidMonth) {
		t.Fatalf("expected ErrInvalidMonth, got %v", err)
	}
}

func TestCompare_DefaultsComparisonToPreviousMonth(t *testing.T) {
	stats := &mockStatRepo{byMonth: map[string][]domainStat.CategoryAmount{
		"2026-08": {{CategoryID: "cat_food", Amount: 5000}},
	}}
	uc := reportapp.NewCompareUseCase(stats, &mockCategoryRepo{}, nil)
	res, err := uc.Execute(context.Background(), reportapp.CompareInput{
		UserID:       "usr_1",
		Type:         "expense",
		PrimaryMonth: "2026-08",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ComparisonPeriod != "2026-07" {
		t.Errorf("expected comparison defaulted to 2026-07, got %s", res.ComparisonPeriod)
	}
}
