package category

import (
	"context"

	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type DeleteCategoryUseCase struct {
	repo domainCategory.Repository
}

func NewDeleteCategoryUseCase(repo domainCategory.Repository) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{repo: repo}
}

func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, id string) error {
	_, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	return uc.repo.Delete(ctx, id)
}
