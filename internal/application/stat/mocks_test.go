package stat_test

import (
	"context"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	"github.com/Dodul01/fundly/internal/domain/category"
	domainStat "github.com/Dodul01/fundly/internal/domain/stat"
)

type mockStatRepo struct {
	breakdown []domainStat.CategoryAmount
	totals    domainStat.Totals
}

func (m *mockStatRepo) BreakdownByCategory(_ context.Context, _, _, _, _ string) ([]domainStat.CategoryAmount, error) {
	return m.breakdown, nil
}

func (m *mockStatRepo) MonthlyTotals(_ context.Context, _, _ string) (domainStat.Totals, error) {
	return m.totals, nil
}

type mockCategoryRepo struct {
	categories map[string]*category.Category
}

func (m *mockCategoryRepo) GetAll(_ context.Context) ([]*category.Category, error) { return nil, nil }
func (m *mockCategoryRepo) GetByType(_ context.Context, _ category.CategoryType) ([]*category.Category, error) {
	return nil, nil
}
func (m *mockCategoryRepo) GetByID(_ context.Context, id string) (*category.Category, error) {
	if m.categories == nil {
		return nil, category.ErrCategoryNotFound
	}
	c, ok := m.categories[id]
	if !ok {
		return nil, category.ErrCategoryNotFound
	}
	return c, nil
}
func (m *mockCategoryRepo) Create(_ context.Context, _ *category.Category) error { return nil }
func (m *mockCategoryRepo) Update(_ context.Context, _ *category.Category) error { return nil }
func (m *mockCategoryRepo) Delete(_ context.Context, _ string) error            { return nil }

type mockBudgetRepo struct {
	budgets []*domainBudget.Budget
}

func (m *mockBudgetRepo) Create(_ context.Context, _ *domainBudget.Budget) error { return nil }
func (m *mockBudgetRepo) GetByID(_ context.Context, _ string) (*domainBudget.Budget, error) {
	return nil, domainBudget.ErrBudgetNotFound
}
func (m *mockBudgetRepo) List(_ context.Context, _, _, _ string) ([]*domainBudget.Budget, error) {
	return m.budgets, nil
}
func (m *mockBudgetRepo) Update(_ context.Context, _ *domainBudget.Budget) error { return nil }
func (m *mockBudgetRepo) Delete(_ context.Context, _ string) error              { return nil }
