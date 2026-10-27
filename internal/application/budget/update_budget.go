package budget

import (
	"context"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
)

type UpdateBudgetInput struct {
	UserID      string  `json:"-"`
	ID          string  `json:"-"`
	CategoryID  *string `json:"category_id"`
	Month       *string `json:"month"`
	LimitAmount *int64  `json:"limit_amount"`
	Note        *string `json:"note"`
	Description *string `json:"description"`
	Expenditure *int64  `json:"expenditure"`
}

type UpdateBudgetUseCase struct {
	repo  domainBudget.Repository
	cats  domainCategory.Repository
	spent SpentProvider
}

func NewUpdateBudgetUseCase(repo domainBudget.Repository, cats domainCategory.Repository, spent SpentProvider) *UpdateBudgetUseCase {
	return &UpdateBudgetUseCase{repo: repo, cats: cats, spent: spent}
}

func (uc *UpdateBudgetUseCase) Execute(ctx context.Context, input UpdateBudgetInput) (*BudgetResponse, error) {
	bg, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}
	if bg.UserID != input.UserID {
		return nil, domainBudget.ErrBudgetNotFound
	}

	if err := bg.Update(input.CategoryID, input.Month, input.LimitAmount, input.Note, input.Description, input.Expenditure); err != nil {
		return nil, err
	}

	if err := uc.repo.Update(ctx, bg); err != nil {
		return nil, err
	}

	res := builder{categories: uc.cats, spent: uc.spent}.toResponse(ctx, bg)
	return &res, nil
}
