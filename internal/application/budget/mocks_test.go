package budget_test

import (
	"context"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	"github.com/Dodul01/fundly/internal/domain/category"
)

type mockBudgetRepo struct {
	budgets   map[string]*domainBudget.Budget
	createErr error
}

func newMockBudgetRepo() *mockBudgetRepo {
	return &mockBudgetRepo{budgets: make(map[string]*domainBudget.Budget)}
}

func (m *mockBudgetRepo) Create(_ context.Context, b *domainBudget.Budget) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.budgets[b.ID] = b
	return nil
}

func (m *mockBudgetRepo) GetByID(_ context.Context, id string) (*domainBudget.Budget, error) {
	b, ok := m.budgets[id]
	if !ok {
		return nil, domainBudget.ErrBudgetNotFound
	}
	return b, nil
}

func (m *mockBudgetRepo) List(_ context.Context, userID, month, categoryID string) ([]*domainBudget.Budget, error) {
	var out []*domainBudget.Budget
	for _, b := range m.budgets {
		if b.UserID != userID {
			continue
		}
		if month != "" && b.Month != month {
			continue
		}
		if categoryID != "" && b.CategoryID != categoryID {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

func (m *mockBudgetRepo) Update(_ context.Context, b *domainBudget.Budget) error {
	if _, ok := m.budgets[b.ID]; !ok {
		return domainBudget.ErrBudgetNotFound
	}
	m.budgets[b.ID] = b
	return nil
}

func (m *mockBudgetRepo) Delete(_ context.Context, id string) error {
	if _, ok := m.budgets[id]; !ok {
		return domainBudget.ErrBudgetNotFound
	}
	delete(m.budgets, id)
	return nil
}

// mockCategoryRepo provides category lookups for enrichment.
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

// mockSpent returns a fixed spent amount for any lookup.
type mockSpent struct {
	amount int64
}

func (m mockSpent) SpentByCategoryMonth(_ context.Context, _, _, _ string) (int64, error) {
	return m.amount, nil
}
