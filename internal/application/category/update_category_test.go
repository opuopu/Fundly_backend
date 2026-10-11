package category_test

import (
	"context"
	"errors"
	"testing"

	appcategory "github.com/Dodul01/fundly/internal/application/category"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

func TestUpdateCategory_Success(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	cat, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[cat.ID] = cat

	uc := appcategory.NewUpdateCategoryUseCase(repo)

	newName := "Pharmacy & Medicine"
	newColor := "#C2185B"
	newIcon := "pills"

	res, err := uc.Execute(context.Background(), appcategory.UpdateCategoryInput{
		ID:    "cat_01",
		Name:  &newName,
		Color: &newColor,
		Icon:  &newIcon,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID != "cat_01" {
		t.Errorf("expected id 'cat_01', got '%s'", res.ID)
	}
	if res.Name != "Pharmacy & Medicine" {
		t.Errorf("expected name 'Pharmacy & Medicine', got '%s'", res.Name)
	}
	if res.Color != "#C2185B" {
		t.Errorf("expected color '#C2185B', got '%s'", res.Color)
	}
	if res.Icon != "pills" {
		t.Errorf("expected icon 'pills', got '%s'", res.Icon)
	}
	if res.Type != "expense" {
		t.Errorf("expected type 'expense', got '%s'", res.Type)
	}

	// Verify in repository
	saved := repo.categories["cat_01"]
	if saved.Name != "Pharmacy & Medicine" || saved.Color != "#C2185B" || saved.Icon != "pills" {
		t.Errorf("repository category not updated properly: %+v", saved)
	}
}

func TestUpdateCategory_PartialUpdate(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	cat, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[cat.ID] = cat

	uc := appcategory.NewUpdateCategoryUseCase(repo)

	newName := "Updated Name Only"
	res, err := uc.Execute(context.Background(), appcategory.UpdateCategoryInput{
		ID:   "cat_01",
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Name != "Updated Name Only" {
		t.Errorf("expected name 'Updated Name Only', got '%s'", res.Name)
	}
	if res.Color != "#000000" {
		t.Errorf("expected color '#000000' preserved, got '%s'", res.Color)
	}
	if res.Icon != "old-icon" {
		t.Errorf("expected icon 'old-icon' preserved, got '%s'", res.Icon)
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewUpdateCategoryUseCase(repo)

	newName := "New Name"
	_, err := uc.Execute(context.Background(), appcategory.UpdateCategoryInput{
		ID:   "cat_nonexistent",
		Name: &newName,
	})
	if !errors.Is(err, domainCategory.ErrCategoryNotFound) {
		t.Fatalf("expected ErrCategoryNotFound, got %v", err)
	}
}

func TestUpdateCategory_EmptyName(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	cat, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[cat.ID] = cat

	uc := appcategory.NewUpdateCategoryUseCase(repo)

	emptyName := "   "
	_, err := uc.Execute(context.Background(), appcategory.UpdateCategoryInput{
		ID:   "cat_01",
		Name: &emptyName,
	})
	if !errors.Is(err, domainCategory.ErrInvalidCategoryName) {
		t.Fatalf("expected ErrInvalidCategoryName, got %v", err)
	}
}

func TestUpdateCategory_NoFieldsToUpdate(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	cat, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[cat.ID] = cat

	uc := appcategory.NewUpdateCategoryUseCase(repo)

	_, err := uc.Execute(context.Background(), appcategory.UpdateCategoryInput{
		ID: "cat_01",
	})
	if !errors.Is(err, domainCategory.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}
}

func TestUpdateCategory_RepoUpdateError(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	cat, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[cat.ID] = cat

	repoErr := errors.New("db error")
	repo.err = repoErr

	uc := appcategory.NewUpdateCategoryUseCase(repo)

	newName := "New Name"
	_, err := uc.Execute(context.Background(), appcategory.UpdateCategoryInput{
		ID:   "cat_01",
		Name: &newName,
	})
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
