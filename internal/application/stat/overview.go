package stat

import (
	"context"
	"math"
	"strings"

	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	domainStat "github.com/Dodul01/fundly/internal/domain/stat"
)

type OverviewResponse struct {
	Month             string  `json:"month"`
	TotalIncome       int64   `json:"total_income"`
	TotalExpense      int64   `json:"total_expense"`
	NetBalance        int64   `json:"net_balance"`
	Currency          string  `json:"currency"`
	BudgetLimit       int64   `json:"budget_limit"`
	BudgetUtilization float64 `json:"budget_utilization"`
}

type OverviewUseCase struct {
	stats   domainStat.Repository
	budgets domainBudget.Repository
	configs ConfigReader
}

func NewOverviewUseCase(stats domainStat.Repository, budgets domainBudget.Repository, configs ConfigReader) *OverviewUseCase {
	return &OverviewUseCase{stats: stats, budgets: budgets, configs: configs}
}

func (uc *OverviewUseCase) Execute(ctx context.Context, userID, month string) (*OverviewResponse, error) {
	month = strings.TrimSpace(month)
	if month == "" {
		month = CurrentMonth()
	}

	totals, err := uc.stats.MonthlyTotals(ctx, userID, month)
	if err != nil {
		return nil, err
	}

	res := &OverviewResponse{
		Month:        month,
		TotalIncome:  totals.TotalIncome,
		TotalExpense: totals.TotalExpense,
		NetBalance:   totals.TotalIncome - totals.TotalExpense,
		Currency:     resolveCurrency(ctx, uc.configs, userID),
	}

	// Budget utilization: this month's total expense against the sum of all budget limits.
	if uc.budgets != nil {
		budgets, err := uc.budgets.List(ctx, userID, month, "")
		if err != nil {
			return nil, err
		}
		var limit int64
		for _, b := range budgets {
			limit += b.LimitAmount
		}
		res.BudgetLimit = limit
		if limit > 0 {
			res.BudgetUtilization = math.Round(float64(totals.TotalExpense)/float64(limit)*10000) / 100
		}
	}

	return res, nil
}
