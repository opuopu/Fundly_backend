package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Dodul01/fundly/internal/domain/budget"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const budgetColumns = `id, user_id, category_id, month, limit_amount, currency, note, description, created_at, updated_at`

type BudgetRepository struct {
	pool *pgxpool.Pool
}

func NewBudgetRepository(pool *pgxpool.Pool) *BudgetRepository {
	return &BudgetRepository{pool: pool}
}

func scanBudget(row pgx.Row) (*budget.Budget, error) {
	var (
		b     budget.Budget
		note  *string
		descr *string
	)
	if err := row.Scan(
		&b.ID,
		&b.UserID,
		&b.CategoryID,
		&b.Month,
		&b.LimitAmount,
		&b.Currency,
		&note,
		&descr,
		&b.CreatedAt,
		&b.UpdatedAt,
	); err != nil {
		return nil, err
	}
	if note != nil {
		b.Note = *note
	}
	if descr != nil {
		b.Description = *descr
	}
	return &b, nil
}

func (r *BudgetRepository) Create(ctx context.Context, b *budget.Budget) error {
	query := `
		INSERT INTO budgets (id, user_id, category_id, month, limit_amount, currency, note, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(ctx, query,
		b.ID, b.UserID, b.CategoryID, b.Month, b.LimitAmount, b.Currency,
		nullIfEmpty(b.Note), nullIfEmpty(b.Description), b.CreatedAt, b.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uq_user_category_month") {
			return budget.ErrBudgetExists
		}
		return fmt.Errorf("inserting budget: %w", err)
	}
	return nil
}

func (r *BudgetRepository) GetByID(ctx context.Context, id string) (*budget.Budget, error) {
	query := `SELECT ` + budgetColumns + ` FROM budgets WHERE id = $1`
	b, err := scanBudget(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, budget.ErrBudgetNotFound
		}
		return nil, fmt.Errorf("querying budget by id: %w", err)
	}
	return b, nil
}

func (r *BudgetRepository) List(ctx context.Context, userID, month, categoryID string) ([]*budget.Budget, error) {
	conditions := []string{"user_id = $1"}
	args := []any{userID}

	if month != "" {
		args = append(args, month)
		conditions = append(conditions, fmt.Sprintf("month = $%d", len(args)))
	}
	if categoryID != "" {
		args = append(args, categoryID)
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", len(args)))
	}

	query := `SELECT ` + budgetColumns + ` FROM budgets WHERE ` +
		strings.Join(conditions, " AND ") + ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying budgets: %w", err)
	}
	defer rows.Close()

	var budgets []*budget.Budget
	for rows.Next() {
		b, err := scanBudget(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning budget row: %w", err)
		}
		budgets = append(budgets, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating budget rows: %w", err)
	}
	return budgets, nil
}

func (r *BudgetRepository) Update(ctx context.Context, b *budget.Budget) error {
	query := `
		UPDATE budgets
		SET category_id = $2, month = $3, limit_amount = $4, note = $5, description = $6, updated_at = $7
		WHERE id = $1
	`
	cmdTag, err := r.pool.Exec(ctx, query,
		b.ID, b.CategoryID, b.Month, b.LimitAmount, nullIfEmpty(b.Note), nullIfEmpty(b.Description), b.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "uq_user_category_month") {
			return budget.ErrBudgetExists
		}
		return fmt.Errorf("updating budget: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return budget.ErrBudgetNotFound
	}
	return nil
}

func (r *BudgetRepository) Delete(ctx context.Context, id string) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM budgets WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting budget: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return budget.ErrBudgetNotFound
	}
	return nil
}
