package category_test

import (
	"context"
	"errors"
	"testing"

	appcategory "github.com/Dodul01/fundly/internal/application/category"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type mockCategoryRepo struct {
	categories map[string]*domainCategory.Category
	err        error
}

func (m *mockCategoryRepo) GetAll(ctx context.Context) ([]*domainCategory.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainCategory.Category
	for _, c := range m.categories {
		res = append(res, c)
	}
	return res, nil
}

func (m *mockCategoryRepo) GetByType(ctx context.Context, catType domainCategory.CategoryType) ([]*domainCategory.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainCategory.Category
	for _, c := range m.categories {
		if c.Type == catType {
			res = append(res, c)
		}
	}
	return res, nil
}

func (m *mockCategoryRepo) GetByID(ctx context.Context, id string) (*domainCategory.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.categories[id]
	if !ok {
		return nil, domainCategory.ErrCategoryNotFound
	}
	return c, nil
}

func (m *mockCategoryRepo) Create(ctx context.Context, cat *domainCategory.Category) error {
	if m.err != nil {
		return m.err
	}
	m.categories[cat.ID] = cat
	return nil
}

func (m *mockCategoryRepo) Update(ctx context.Context, cat *domainCategory.Category) error {
	if m.err != nil {
		return m.err
	}
	m.categories[cat.ID] = cat
	return nil
}

func (m *mockCategoryRepo) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.categories, id)
	return nil
}

func TestListCategoriesUseCase_All(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_1", "Food", "EXPENSE", "bowl", "#FF0000")
	c2, _ := domainCategory.NewCategory("cat_2", "Salary", "INCOME", "wallet", "#00FF00")
	repo.categories[c1.ID] = c1
	repo.categories[c2.ID] = c2

	uc := appcategory.NewListCategoriesUseCase(repo)

	res, err := uc.Execute(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(res))
	}
}

func TestListCategoriesUseCase_FilterExpense(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_1", "Food", "EXPENSE", "bowl", "#FF0000")
	c2, _ := domainCategory.NewCategory("cat_2", "Salary", "INCOME", "wallet", "#00FF00")
	repo.categories[c1.ID] = c1
	repo.categories[c2.ID] = c2

	uc := appcategory.NewListCategoriesUseCase(repo)

	res, err := uc.Execute(context.Background(), "expense")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 category, got %d", len(res))
	}
	if res[0].Type != "expense" {
		t.Errorf("expected type 'expense', got '%s'", res[0].Type)
	}
}

func TestListCategoriesUseCase_FilterIncome(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_1", "Food", "EXPENSE", "bowl", "#FF0000")
	c2, _ := domainCategory.NewCategory("cat_2", "Salary", "INCOME", "wallet", "#00FF00")
	repo.categories[c1.ID] = c1
	repo.categories[c2.ID] = c2

	uc := appcategory.NewListCategoriesUseCase(repo)

	res, err := uc.Execute(context.Background(), "INCOME")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res) != 1 {
		t.Fatalf("expected 1 category, got %d", len(res))
	}
	if res[0].Type != "income" {
		t.Errorf("expected type 'income', got '%s'", res[0].Type)
	}
}

func TestListCategoriesUseCase_InvalidFilter(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewListCategoriesUseCase(repo)

	_, err := uc.Execute(context.Background(), "invalid_type")
	if !errors.Is(err, domainCategory.ErrInvalidCategoryType) {
		t.Fatalf("expected ErrInvalidCategoryType, got %v", err)
	}
}

func TestListCategoriesUseCase_Empty(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewListCategoriesUseCase(repo)

	res, err := uc.Execute(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res == nil || len(res) != 0 {
		t.Errorf("expected non-nil empty slice, got %+v", res)
	}
}

func TestListCategoriesUseCase_RepoError(t *testing.T) {
	repoErr := errors.New("db query error")
	repo := &mockCategoryRepo{
		categories: make(map[string]*domainCategory.Category),
		err:        repoErr,
	}
	uc := appcategory.NewListCategoriesUseCase(repo)

	_, err := uc.Execute(context.Background(), "")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
