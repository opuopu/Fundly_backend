package category

import (
	"context"

	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type UpdateCategoryInput struct {
	ID    string  `json:"id"`
	Name  *string `json:"name"`
	Color *string `json:"color"`
	Icon  *string `json:"icon"`
}

type UpdateCategoryUseCase struct {
	repo domainCategory.Repository
}

func NewUpdateCategoryUseCase(repo domainCategory.Repository) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{repo: repo}
}

func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, input UpdateCategoryInput) (*CategoryResponse, error) {
	cat, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	if err := cat.Update(input.Name, input.Color, input.Icon); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, cat); err != nil {
		return nil, err
	}

	res := ToCategoryResponse(cat)
	return &res, nil
}
