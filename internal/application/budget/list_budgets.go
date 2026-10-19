package budget

import (
	"context"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type ListBudgetsUseCase struct {
	repo  domainBudget.Repository
	cats  domainCategory.Repository
	spent SpentProvider
}

func NewListBudgetsUseCase(repo domainBudget.Repository, cats domainCategory.Repository, spent SpentProvider) *ListBudgetsUseCase {
	return &ListBudgetsUseCase{repo: repo, cats: cats, spent: spent}
}

func (uc *ListBudgetsUseCase) Execute(ctx context.Context, userID, month, categoryID string) ([]BudgetResponse, error) {
	budgets, err := uc.repo.List(ctx, userID, month, categoryID)
	if err != nil {
		return nil, err
	}

	b := builder{categories: uc.cats, spent: uc.spent}
	res := make([]BudgetResponse, 0, len(budgets))
	for _, bg := range budgets {
		res = append(res, b.toResponse(ctx, bg))
	}
	return res, nil
}
