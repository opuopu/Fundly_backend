package budget

import "context"

type Repository interface {
	Create(ctx context.Context, b *Budget) error
	GetByID(ctx context.Context, id string) (*Budget, error)
	List(ctx context.Context, userID, month, categoryID string) ([]*Budget, error)
	Update(ctx context.Context, b *Budget) error
	Delete(ctx context.Context, id string) error
}
