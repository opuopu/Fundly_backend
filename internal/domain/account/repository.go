package account

import "context"

type Repository interface {
	GetByID(ctx context.Context, id string) (*Account, error)
	GetByUserID(ctx context.Context, userID string) ([]*Account, error)
	GetByUserIDAndGroup(ctx context.Context, userID, group string) ([]*Account, error)
	Create(ctx context.Context, acc *Account) error
	Update(ctx context.Context, acc *Account) error
	Delete(ctx context.Context, id string) error
}
