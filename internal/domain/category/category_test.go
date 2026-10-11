package category_test

import (
	"errors"
	"testing"

	"github.com/Dodul01/fundly/internal/domain/category"
)

func TestNewCategory_Success(t *testing.T) {
	cat, err := category.NewCategory("cat_01", "Groceries", "EXPENSE", "cart", "#FF5722")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cat.ID != "cat_01" || cat.Name != "Groceries" || cat.Type != category.TypeExpense || cat.Icon != "cart" || cat.Color != "#FF5722" {
		t.Errorf("unexpected category fields: %+v", cat)
	}
}

func TestNewCategory_Defaults(t *testing.T) {
	cat, err := category.NewCategory("cat_02", "Salary", "income", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cat.Type != category.TypeIncome {
		t.Errorf("expected TypeIncome, got %s", cat.Type)
	}
	if cat.Icon != "tag" {
		t.Errorf("expected default icon 'tag', got '%s'", cat.Icon)
	}
	if cat.Color != "#2196F3" {
		t.Errorf("expected default color '#2196F3', got '%s'", cat.Color)
	}
}

func TestNewCategory_InvalidName(t *testing.T) {
	_, err := category.NewCategory("cat_03", "   ", "EXPENSE", "", "")
	if !errors.Is(err, category.ErrInvalidCategoryName) {
		t.Fatalf("expected ErrInvalidCategoryName, got %v", err)
	}
}

func TestNewCategory_InvalidType(t *testing.T) {
	_, err := category.NewCategory("cat_04", "Test", "transfer", "", "")
	if !errors.Is(err, category.ErrInvalidCategoryType) {
		t.Fatalf("expected ErrInvalidCategoryType, got %v", err)
	}
}

func TestCategory_Update(t *testing.T) {
	cat, _ := category.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")

	newName := "New Name"
	newColor := "#FFFFFF"
	newIcon := "new-icon"
	err := cat.Update(&newName, &newColor, &newIcon)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cat.Name != "New Name" || cat.Color != "#FFFFFF" || cat.Icon != "new-icon" {
		t.Errorf("unexpected updated fields: %+v", cat)
	}

	// Empty name error
	emptyName := "  "
	err = cat.Update(&emptyName, nil, nil)
	if !errors.Is(err, category.ErrInvalidCategoryName) {
		t.Fatalf("expected ErrInvalidCategoryName, got %v", err)
	}

	// No fields error
	err = cat.Update(nil, nil, nil)
	if !errors.Is(err, category.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}
}
