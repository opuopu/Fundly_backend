package postgres

import (
	"context"

	"github.com/Dodul01/fundly/internal/domain/account"
)

type AccountRepository struct {
	// db connection/pool (e.g. *sql.DB or *pgxpool.Pool) will be injected here
}

func NewAccountRepository() *AccountRepository {
	return &AccountRepository{}
}

func (r *AccountRepository) GetByID(ctx context.Context, id string) (*account.Account, error) {
	// TODO: Replace with database query
	return nil, account.ErrAccountNotFound
}

func (r *AccountRepository) GetByUserID(ctx context.Context, userID string) ([]*account.Account, error) {
	// TODO: Replace with database query
	return nil, nil
}

func (r *AccountRepository) Create(ctx context.Context, acc *account.Account) error {
	// TODO: Replace with database insert
	return nil
}

func (r *AccountRepository) Update(ctx context.Context, acc *account.Account) error {
	// TODO: Replace with database update
	return nil
}
