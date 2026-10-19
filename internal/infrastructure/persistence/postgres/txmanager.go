package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// querier is the subset of methods shared by *pgxpool.Pool and pgx.Tx, letting repositories
// run their statements against either a pooled connection or an active transaction.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type txCtxKey struct{}

// ctxWithQuerier stashes an active transaction on the context.
func ctxWithQuerier(ctx context.Context, q querier) context.Context {
	return context.WithValue(ctx, txCtxKey{}, q)
}

// dbFromContext returns the transaction bound to the context, or the fallback pool when the
// call is not running inside a managed transaction.
func dbFromContext(ctx context.Context, fallback querier) querier {
	if q, ok := ctx.Value(txCtxKey{}).(querier); ok && q != nil {
		return q
	}
	return fallback
}

// TxManager runs a function within a single database transaction. Repositories that read the
// querier from the context automatically participate in that transaction, so a failure
// anywhere in the function rolls back every write.
type TxManager struct {
	pool *pgxpool.Pool
}

func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

func (m *TxManager) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}

	if err := fn(ctxWithQuerier(ctx, tx)); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}
	return nil
}
