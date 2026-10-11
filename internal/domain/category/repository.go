package category

import "context"

type Repository interface {
	GetAll(ctx context.Context) ([]*Category, error)
	GetByType(ctx context.Context, catType CategoryType) ([]*Category, error)
	GetByID(ctx context.Context, id string) (*Category, error)
	Create(ctx context.Context, cat *Category) error
	Update(ctx context.Context, cat *Category) error
	Delete(ctx context.Context, id string) error
}
