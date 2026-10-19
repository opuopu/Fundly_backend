package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/Dodul01/fundly/internal/domain/stat"
	"github.com/jackc/pgx/v5/pgxpool"
)

type StatRepository struct {
	pool *pgxpool.Pool
}

func NewStatRepository(pool *pgxpool.Pool) *StatRepository {
	return &StatRepository{pool: pool}
}

func (r *StatRepository) BreakdownByCategory(ctx context.Context, userID, txType, month, accountID string) ([]stat.CategoryAmount, error) {
	conditions := []string{
		"user_id = $1",
		"type = $2",
		"to_char(transaction_date, 'YYYY-MM') = $3",
		"category_id IS NOT NULL",
	}
	args := []any{userID, strings.ToUpper(txType), month}

	if accountID != "" {
		args = append(args, accountID)
		conditions = append(conditions, fmt.Sprintf("(from_account_id = $%d OR to_account_id = $%d)", len(args), len(args)))
	}

	query := `
		SELECT category_id, COALESCE(SUM(amount), 0) AS total
		FROM transactions
		WHERE ` + strings.Join(conditions, " AND ") + `
		GROUP BY category_id
		ORDER BY total DESC
	`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying category breakdown: %w", err)
	}
	defer rows.Close()

	var out []stat.CategoryAmount
	for rows.Next() {
		var ca stat.CategoryAmount
		if err := rows.Scan(&ca.CategoryID, &ca.Amount); err != nil {
			return nil, fmt.Errorf("scanning breakdown row: %w", err)
		}
		out = append(out, ca)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating breakdown rows: %w", err)
	}
	return out, nil
}

func (r *StatRepository) MonthlyTotals(ctx context.Context, userID, month string) (stat.Totals, error) {
	query := `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = 'INCOME'), 0),
			COALESCE(SUM(amount) FILTER (WHERE type = 'EXPENSE'), 0)
		FROM transactions
		WHERE user_id = $1 AND to_char(transaction_date, 'YYYY-MM') = $2
	`

	var t stat.Totals
	if err := r.pool.QueryRow(ctx, query, userID, month).Scan(&t.TotalIncome, &t.TotalExpense); err != nil {
		return stat.Totals{}, fmt.Errorf("querying monthly totals: %w", err)
	}
	return t, nil
}
