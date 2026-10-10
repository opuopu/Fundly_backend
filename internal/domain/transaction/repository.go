package transaction

import "context"

type Repository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	ListByAccountID(ctx context.Context, accountID string) ([]*Transaction, error)
}
