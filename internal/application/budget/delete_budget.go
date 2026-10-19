package budget

import (
	"context"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
)

type DeleteBudgetUseCase struct {
	repo domainBudget.Repository
}

func NewDeleteBudgetUseCase(repo domainBudget.Repository) *DeleteBudgetUseCase {
	return &DeleteBudgetUseCase{repo: repo}
}

func (uc *DeleteBudgetUseCase) Execute(ctx context.Context, userID, id string) error {
	bg, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if bg.UserID != userID {
		return domainBudget.ErrBudgetNotFound
	}

	return uc.repo.Delete(ctx, id)
}
