package account

import (
	"context"
	"strings"

	domainAccount "github.com/Dodul01/fundly/internal/domain/account"
)

type AccountsSummaryResponse struct {
	TotalBalance int64  `json:"total_balance"`
	TotalCash    int64  `json:"total_cash"`
	TotalCard    int64  `json:"total_card"`
	TotalSavings int64  `json:"total_savings"`
	Currency     string `json:"currency"`
}

type GetAccountsSummaryUseCase struct {
	repo domainAccount.Repository
}

func NewGetAccountsSummaryUseCase(repo domainAccount.Repository) *GetAccountsSummaryUseCase {
	return &GetAccountsSummaryUseCase{repo: repo}
}

func (uc *GetAccountsSummaryUseCase) Execute(ctx context.Context, userID string) (*AccountsSummaryResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, domainAccount.ErrAccountNotFound
	}

	accounts, err := uc.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := &AccountsSummaryResponse{
		Currency: "USD",
	}

	for _, acc := range accounts {
		res.TotalBalance += acc.Balance
		switch strings.ToLower(acc.Group) {
		case "cash":
			res.TotalCash += acc.Balance
		case "card", "checking":
			res.TotalCard += acc.Balance
		case "savings":
			res.TotalSavings += acc.Balance
		}
		if acc.Currency != "" {
			res.Currency = acc.Currency
		}
	}

	return res, nil
}
