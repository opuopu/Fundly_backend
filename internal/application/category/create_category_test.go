package category_test

import (
	"context"
	"errors"
	"testing"

	appcategory "github.com/Dodul01/fundly/internal/application/category"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

func TestCreateCategory_Success(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewCreateCategoryUseCase(repo)

	input := appcategory.CreateCategoryInput{
		Name:  "Medicine",
		Type:  "expense",
		Icon:  "pills",
		Color: "#E91E63",
	}

	res, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.ID == "" {
		t.Errorf("expected generated category ID")
	}
	if res.Name != "Medicine" {
		t.Errorf("expected name 'Medicine', got '%s'", res.Name)
	}
	if res.Type != "expense" {
		t.Errorf("expected type 'expense', got '%s'", res.Type)
	}
	if res.Icon != "pills" {
		t.Errorf("expected icon 'pills', got '%s'", res.Icon)
	}
	if res.Color != "#E91E63" {
		t.Errorf("expected color '#E91E63', got '%s'", res.Color)
	}

	// Verify persistence in repo
	saved, ok := repo.categories[res.ID]
	if !ok {
		t.Fatalf("expected category to be stored in repo")
	}
	if saved.Name != "Medicine" || saved.Type != domainCategory.TypeExpense {
		t.Errorf("saved category does not match: %+v", saved)
	}
}

func TestCreateCategory_Defaults(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewCreateCategoryUseCase(repo)

	input := appcategory.CreateCategoryInput{
		Name: "Bonus",
		Type: "income",
	}

	res, err := uc.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Icon != "tag" {
		t.Errorf("expected default icon 'tag', got '%s'", res.Icon)
	}
	if res.Color != "#2196F3" {
		t.Errorf("expected default color '#2196F3', got '%s'", res.Color)
	}
}

func TestCreateCategory_InvalidName(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewCreateCategoryUseCase(repo)

	input := appcategory.CreateCategoryInput{
		Name: "   ",
		Type: "expense",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainCategory.ErrInvalidCategoryName) {
		t.Fatalf("expected ErrInvalidCategoryName, got %v", err)
	}
}

func TestCreateCategory_InvalidType(t *testing.T) {
	repo := &mockCategoryRepo{categories: make(map[string]*domainCategory.Category)}
	uc := appcategory.NewCreateCategoryUseCase(repo)

	input := appcategory.CreateCategoryInput{
		Name: "Test",
		Type: "invalid_type",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, domainCategory.ErrInvalidCategoryType) {
		t.Fatalf("expected ErrInvalidCategoryType, got %v", err)
	}
}

func TestCreateCategory_RepoError(t *testing.T) {
	repoErr := errors.New("db insert failure")
	repo := &mockCategoryRepo{
		categories: make(map[string]*domainCategory.Category),
		err:        repoErr,
	}
	uc := appcategory.NewCreateCategoryUseCase(repo)

	input := appcategory.CreateCategoryInput{
		Name: "Test",
		Type: "expense",
	}

	_, err := uc.Execute(context.Background(), input)
	if !errors.Is(err, repoErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
