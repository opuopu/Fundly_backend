package category

import (
	"context"
	"strings"

	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	"github.com/google/uuid"
)

type CreateCategoryInput struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type CreateCategoryUseCase struct {
	repo domainCategory.Repository
}

func NewCreateCategoryUseCase(repo domainCategory.Repository) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{repo: repo}
}

func (uc *CreateCategoryUseCase) Execute(ctx context.Context, input CreateCategoryInput) (*CategoryResponse, error) {
	categoryID := "cat_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]

	cat, err := domainCategory.NewCategory(
		categoryID,
		input.Name,
		input.Type,
		input.Icon,
		input.Color,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, cat); err != nil {
		return nil, err
	}

	res := ToCategoryResponse(cat)
	return &res, nil
}
