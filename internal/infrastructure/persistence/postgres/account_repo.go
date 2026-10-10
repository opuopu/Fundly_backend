package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/account"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AccountRepository struct {
	pool *pgxpool.Pool
}

func NewAccountRepository(pool *pgxpool.Pool) *AccountRepository {
	return &AccountRepository{pool: pool}
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*account.Account, error) {
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var acc account.Account
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.Currency,
		&acc.Balance,
		&acc.CreatedAt,
		&acc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, account.ErrAccountNotFound
		}
		return nil, fmt.Errorf("querying account by id: %w", err)
	}

	return &acc, nil
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID string) ([]*account.Account, error) {
	query := `
		SELECT id, user_id, currency, balance, created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying accounts by user id: %w", err)
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		var acc account.Account
		if err := rows.Scan(
			&acc.ID,
			&acc.UserID,
			&acc.Currency,
			&acc.Balance,
			&acc.CreatedAt,
			&acc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning account row: %w", err)
		}
		accounts = append(accounts, &acc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating account rows: %w", err)
	}

	return accounts, nil
}

func (r *AccountRepository) Create(ctx context.Context, acc *account.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, currency, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		acc.ID,
		acc.UserID,
		acc.Currency,
		acc.Balance,
		acc.CreatedAt,
		acc.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting account: %w", err)
	}

	return nil
}

func (r *AccountRepository) Update(ctx context.Context, acc *account.Account) error {
	query := `
		UPDATE accounts
		SET currency = $2, balance = $3, updated_at = $4
		WHERE id = $1
	`

	cmdTag, err := r.pool.Exec(ctx, query, acc.ID, acc.Currency, acc.Balance, acc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating account: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return account.ErrAccountNotFound
	}

	return nil
}
