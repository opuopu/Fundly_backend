package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/transaction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTransactionNotFound = errors.New("transaction not found")

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *transaction.Transaction) error {
	query := `
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, currency, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		tx.ID,
		tx.FromAccountID,
		tx.ToAccountID,
		tx.Amount,
		tx.Currency,
		string(tx.Type),
		string(tx.Status),
		tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*transaction.Transaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount, currency, type, status, created_at
		FROM transactions
		WHERE id = $1
	`

	var tx transaction.Transaction
	var txType string
	var txStatus string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&tx.ID,
		&tx.FromAccountID,
		&tx.ToAccountID,
		&tx.Amount,
		&tx.Currency,
		&txType,
		&txStatus,
		&tx.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("querying transaction by id: %w", err)
	}

	tx.Type = transaction.Type(txType)
	tx.Status = transaction.Status(txStatus)

	return &tx, nil
}

func (r *TransactionRepository) ListByAccountID(ctx context.Context, accountID string) ([]*transaction.Transaction, error) {
	query := `
		SELECT id, from_account_id, to_account_id, amount, currency, type, status, created_at
		FROM transactions
		WHERE from_account_id = $1 OR to_account_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("querying transactions by account id: %w", err)
	}
	defer rows.Close()

	var txs []*transaction.Transaction
	for rows.Next() {
		var tx transaction.Transaction
		var txType string
		var txStatus string

		if err := rows.Scan(
			&tx.ID,
			&tx.FromAccountID,
			&tx.ToAccountID,
			&tx.Amount,
			&tx.Currency,
			&txType,
			&txStatus,
			&tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning transaction row: %w", err)
		}

		tx.Type = transaction.Type(txType)
		tx.Status = transaction.Status(txStatus)
		txs = append(txs, &tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction rows: %w", err)
	}

	return txs, nil
}
