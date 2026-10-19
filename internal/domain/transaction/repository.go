package transaction

import "context"

// Filter captures the query parameters supported by the transactions listing endpoint.
type Filter struct {
	UserID     string
	Type       string
	CategoryID string
	AccountID  string
	Date       string // YYYY-MM-DD
	StartDate  string // YYYY-MM-DD
	EndDate    string // YYYY-MM-DD
	Page       int
	Limit      int
}

type Repository interface {
	Create(ctx context.Context, tx *Transaction) error
	GetByID(ctx context.Context, id string) (*Transaction, error)
	ListByAccountID(ctx context.Context, accountID string) ([]*Transaction, error)
	List(ctx context.Context, filter Filter) ([]*Transaction, int, error)
	Update(ctx context.Context, tx *Transaction) error
	Delete(ctx context.Context, id string) error
	Search(ctx context.Context, userID, q string, limit int) ([]*Transaction, error)
}
