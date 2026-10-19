package transaction

import "context"

// Atomic runs fn inside a single database transaction, rolling back on any error. The
// postgres TxManager satisfies it. It may be nil (e.g. in unit tests with in-memory repos),
// in which case fn runs directly without a transaction.
type Atomic interface {
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}

// runAtomic executes fn within a transaction when a is non-nil, otherwise runs it directly.
func runAtomic(ctx context.Context, a Atomic, fn func(ctx context.Context) error) error {
	if a == nil {
		return fn(ctx)
	}
	return a.Do(ctx, fn)
}
