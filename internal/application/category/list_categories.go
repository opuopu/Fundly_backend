package category

import (
	"context"
	"strings"

	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type CategoryResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Type  string `json:"type"` // "income" or "expense"
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

func ToCategoryResponse(cat *domainCategory.Category) CategoryResponse {
	return CategoryResponse{
		ID:    cat.ID,
		Name:  cat.Name,
		Type:  strings.ToLower(string(cat.Type)),
		Icon:  cat.Icon,
		Color: cat.Color,
	}
}

type ListCategoriesUseCase struct {
	repo domainCategory.Repository
}

func NewListCategoriesUseCase(repo domainCategory.Repository) *ListCategoriesUseCase {
	return &ListCategoriesUseCase{repo: repo}
}

func (uc *ListCategoriesUseCase) Execute(ctx context.Context, typeFilter string) ([]CategoryResponse, error) {
	typeFilter = strings.TrimSpace(typeFilter)

	var categories []*domainCategory.Category
	var err error

	if typeFilter != "" {
		catType, parseErr := domainCategory.ParseCategoryType(typeFilter)
		if parseErr != nil {
			return nil, parseErr
		}
		categories, err = uc.repo.GetByType(ctx, catType)
	} else {
		categories, err = uc.repo.GetAll(ctx)
	}

	if err != nil {
		return nil, err
	}

	res := make([]CategoryResponse, 0, len(categories))
	for _, c := range categories {
		res = append(res, ToCategoryResponse(c))
	}

	return res, nil
}
