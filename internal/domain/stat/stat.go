package stat

import (
	"context"
	"time"
)

type AccountStats struct {
	AccountID        string    `json:"account_id"`
	TotalInflow      int64     `json:"total_inflow"`
	TotalOutflow     int64     `json:"total_outflow"`
	TransactionCount int       `json:"transaction_count"`
	CalculatedAt     time.Time `json:"calculated_at"`
}

// CategoryAmount is a single aggregated slice of a category breakdown.
type CategoryAmount struct {
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"`
}

// Totals holds monthly income/expense aggregates for a user.
type Totals struct {
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
}

// Repository exposes read-only aggregation queries over recorded transactions.
type Repository interface {
	// BreakdownByCategory returns per-category sums for a given type (INCOME/EXPENSE),
	// month (YYYY-MM) and optional account filter, ordered by amount descending.
	BreakdownByCategory(ctx context.Context, userID, txType, month, accountID string) ([]CategoryAmount, error)
	// MonthlyTotals returns the income and expense totals for a user in a given month.
	MonthlyTotals(ctx context.Context, userID, month string) (Totals, error)
}
