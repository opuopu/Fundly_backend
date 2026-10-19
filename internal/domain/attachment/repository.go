package attachment

import "context"

type Repository interface {
	Create(ctx context.Context, a *Attachment) error
	GetByID(ctx context.Context, id string) (*Attachment, error)
	Delete(ctx context.Context, id string) error
	ListByTransactionID(ctx context.Context, transactionID string) ([]*Attachment, error)
	CountByTransactionID(ctx context.Context, transactionID string) (int, error)
	// LinkToTransaction associates the given attachment IDs with a transaction for a user,
	// ignoring IDs that do not belong to the user.
	LinkToTransaction(ctx context.Context, userID, transactionID string, attachmentIDs []string) error
}
