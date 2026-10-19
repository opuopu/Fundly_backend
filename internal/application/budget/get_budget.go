package budget

import (
	"context"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type GetBudgetUseCase struct {
	repo  domainBudget.Repository
	cats  domainCategory.Repository
	spent SpentProvider
}

func NewGetBudgetUseCase(repo domainBudget.Repository, cats domainCategory.Repository, spent SpentProvider) *GetBudgetUseCase {
	return &GetBudgetUseCase{repo: repo, cats: cats, spent: spent}
}

func (uc *GetBudgetUseCase) Execute(ctx context.Context, userID, id string) (*BudgetResponse, error) {
	bg, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if bg.UserID != userID {
		return nil, domainBudget.ErrBudgetNotFound
	}

	res := builder{categories: uc.cats, spent: uc.spent}.toResponse(ctx, bg)
	return &res, nil
}
