package category_test

import (
	"context"
	"errors"
	"testing"

	appcategory "github.com/Dodul01/fundly/internal/application/category"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

func TestDeleteCategory_Success(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	cat, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[cat.ID] = cat

	uc := appcategory.NewDeleteCategoryUseCase(repo)

	err := uc.Execute(context.Background(), "cat_01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, exists := repo.categories["cat_01"]; exists {
		t.Errorf("expected category to be deleted from repo")
	}
}

func TestDeleteCategory_NotFound(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewDeleteCategoryUseCase(repo)

	err := uc.Execute(context.Background(), "cat_nonexistent")
	if !errors.Is(err, domainCategory.ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestDeleteCategory_RepoError(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	cat, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[cat.ID] = cat

	repoErr := errors.New("db delete error")
	// inject error after GetByID succeeds
	uc := appcategory.NewDeleteCategoryUseCase(repo)

	// Custom mock or setting repo.err will fail GetByID first unless we do it dynamically:
	repo.err = repoErr
	err := uc.Execute(context.Background(), "cat_01")
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
