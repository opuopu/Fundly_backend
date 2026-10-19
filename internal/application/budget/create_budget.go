package budget

import (
	"context"
	"strings"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	"github.com/google/uuid"
)

type CreateBudgetInput struct {
	UserID      string `json:"-"`
	Month       string `json:"month"`
	LimitAmount int64  `json:"limit_amount"`
	CategoryID  string `json:"category_id"`
	Currency    string `json:"currency"`
	Note        string `json:"note"`
	Description string `json:"description"`
}

type CreateBudgetUseCase struct {
	repo  domainBudget.Repository
	cats  domainCategory.Repository
	spent SpentProvider
}

func NewCreateBudgetUseCase(repo domainBudget.Repository, cats domainCategory.Repository, spent SpentProvider) *CreateBudgetUseCase {
	return &CreateBudgetUseCase{repo: repo, cats: cats, spent: spent}
}

func (uc *CreateBudgetUseCase) Execute(ctx context.Context, input CreateBudgetInput) (*BudgetResponse, error) {
	budgetID := "bg_" + strings.ReplaceAll(uuid.New().String(), "-", "")[:8]

	bg, err := domainBudget.NewBudget(
		budgetID,
		input.UserID,
		input.CategoryID,
		input.Month,
		input.LimitAmount,
		input.Currency,
		input.Note,
		input.Description,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.repo.Create(ctx, bg); err != nil {
		return nil, err
	}

	res := builder{categories: uc.cats, spent: uc.spent}.toResponse(ctx, bg)
	return &res, nil
}
