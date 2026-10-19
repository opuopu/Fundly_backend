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

// db returns the active transaction bound to ctx, or the pool otherwise.
func (r *AccountRepository) db(ctx context.Context) querier {
	return dbFromContext(ctx, r.pool)
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*account.Account, error) {
	query := `
		SELECT id, user_id, name, group_type, currency, balance, COALESCE(description, ''), created_at, updated_at
		FROM accounts
		WHERE id = $1
	`

	var acc account.Account
	err := r.db(ctx).QueryRow(ctx, query, id).Scan(
		&acc.ID,
		&acc.UserID,
		&acc.Name,
		&acc.Group,
		&acc.Currency,
		&acc.Balance,
		&acc.Description,
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
		SELECT id, user_id, name, group_type, currency, balance, COALESCE(description, ''), created_at, updated_at
		FROM accounts
		WHERE user_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db(ctx).Query(ctx, query, userID)
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
			&acc.Name,
			&acc.Group,
			&acc.Currency,
			&acc.Balance,
			&acc.Description,
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

func (r *AccountRepository) GetByUserIDAndGroup(ctx context.Context, userID, group string) ([]*account.Account, error) {
	query := `
		SELECT id, user_id, name, group_type, currency, balance, COALESCE(description, ''), created_at, updated_at
		FROM accounts
		WHERE user_id = $1 AND LOWER(group_type) = LOWER($2)
		ORDER BY created_at ASC
	`

	rows, err := r.db(ctx).Query(ctx, query, userID, group)
	if err != nil {
		return nil, fmt.Errorf("querying accounts by user id and group: %w", err)
	}
	defer rows.Close()

	var accounts []*account.Account
	for rows.Next() {
		var acc account.Account
		if err := rows.Scan(
			&acc.ID,
			&acc.UserID,
			&acc.Name,
			&acc.Group,
			&acc.Currency,
			&acc.Balance,
			&acc.Description,
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
		INSERT INTO accounts (id, user_id, name, group_type, currency, balance, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	var descVal *string
	if acc.Description != "" {
		descVal = &acc.Description
	}

	_, err := r.db(ctx).Exec(
		ctx,
		query,
		acc.ID,
		acc.UserID,
		acc.Name,
		acc.Group,
		acc.Currency,
		acc.Balance,
		descVal,
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
		SET name = $2, group_type = $3, currency = $4, balance = $5, description = $6, updated_at = $7
		WHERE id = $1
	`

	var descVal *string
	if acc.Description != "" {
		descVal = &acc.Description
	}

	cmdTag, err := r.db(ctx).Exec(ctx, query, acc.ID, acc.Name, acc.Group, acc.Currency, acc.Balance, descVal, acc.UpdatedAt)
	if err != nil {
		return fmt.Errorf("updating account: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return account.ErrAccountNotFound
	}

	return nil
}

func (r *AccountRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM accounts
		WHERE id = $1
	`

	cmdTag, err := r.db(ctx).Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting account: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return account.ErrAccountNotFound
	}

	return nil
}
