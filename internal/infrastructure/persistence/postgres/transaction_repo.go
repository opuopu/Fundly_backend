package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/Dodul01/fundly/internal/domain/transaction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrTransactionNotFound is retained as an alias for backwards compatibility; the
// canonical sentinel now lives in the domain package.
var ErrTransactionNotFound = transaction.ErrTransactionNotFound

const txColumns = `id, user_id, from_account_id, to_account_id, category_id, amount, currency, type, status, note, description, transaction_date, created_at`

type TransactionRepository struct {
	pool *pgxpool.Pool
}

func NewTransactionRepository(pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{pool: pool}
}

// db returns the active transaction bound to ctx, or the pool otherwise.
func (r *TransactionRepository) db(ctx context.Context) querier {
	return dbFromContext(ctx, r.pool)
}

// scanRow scans a single transaction row, tolerating NULL account/category columns.
func scanRow(row pgx.Row) (*transaction.Transaction, error) {
	var (
		tx       transaction.Transaction
		fromAcc  sql.NullString
		toAcc    sql.NullString
		catID    sql.NullString
		note     sql.NullString
		descr    sql.NullString
		txType   string
		txStatus string
	)

	if err := row.Scan(
		&tx.ID,
		&tx.UserID,
		&fromAcc,
		&toAcc,
		&catID,
		&tx.Amount,
		&tx.Currency,
		&txType,
		&txStatus,
		&note,
		&descr,
		&tx.Date,
		&tx.CreatedAt,
	); err != nil {
		return nil, err
	}

	tx.FromAccountID = fromAcc.String
	tx.ToAccountID = toAcc.String
	tx.CategoryID = catID.String
	tx.Note = note.String
	tx.Description = descr.String
	tx.Type = transaction.Type(txType)
	tx.Status = transaction.Status(txStatus)

	return &tx, nil
}

func (r *TransactionRepository) Create(ctx context.Context, tx *transaction.Transaction) error {
	query := `
		INSERT INTO transactions (id, user_id, from_account_id, to_account_id, category_id, amount, currency, type, status, note, description, transaction_date, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db(ctx).Exec(
		ctx,
		query,
		tx.ID,
		nullIfEmpty(tx.UserID),
		nullIfEmpty(tx.FromAccountID),
		nullIfEmpty(tx.ToAccountID),
		nullIfEmpty(tx.CategoryID),
		tx.Amount,
		tx.Currency,
		string(tx.Type),
		string(tx.Status),
		nullIfEmpty(tx.Note),
		nullIfEmpty(tx.Description),
		tx.Date,
		tx.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id string) (*transaction.Transaction, error) {
	query := `SELECT ` + txColumns + ` FROM transactions WHERE id = $1`

	tx, err := scanRow(r.db(ctx).QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, transaction.ErrTransactionNotFound
		}
		return nil, fmt.Errorf("querying transaction by id: %w", err)
	}

	return tx, nil
}

func (r *TransactionRepository) ListByAccountID(ctx context.Context, accountID string) ([]*transaction.Transaction, error) {
	query := `SELECT ` + txColumns + ` FROM transactions WHERE from_account_id = $1 OR to_account_id = $1 ORDER BY transaction_date DESC`

	rows, err := r.db(ctx).Query(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("querying transactions by account id: %w", err)
	}
	defer rows.Close()

	return collectRows(rows)
}

func (r *TransactionRepository) List(ctx context.Context, filter transaction.Filter) ([]*transaction.Transaction, int, error) {
	var conditions []string
	var args []any

	add := func(clause string, value any) {
		args = append(args, value)
		conditions = append(conditions, fmt.Sprintf(clause, len(args)))
	}

	if filter.UserID != "" {
		add("user_id = $%d", filter.UserID)
	}
	if filter.Type != "" {
		add("type = $%d", strings.ToUpper(filter.Type))
	}
	if filter.CategoryID != "" {
		add("category_id = $%d", filter.CategoryID)
	}
	if filter.AccountID != "" {
		args = append(args, filter.AccountID)
		conditions = append(conditions, fmt.Sprintf("(from_account_id = $%d OR to_account_id = $%d)", len(args), len(args)))
	}
	if filter.Date != "" {
		add("transaction_date::date = $%d", filter.Date)
	}
	if filter.StartDate != "" {
		add("transaction_date::date >= $%d", filter.StartDate)
	}
	if filter.EndDate != "" {
		add("transaction_date::date <= $%d", filter.EndDate)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	if err := r.db(ctx).QueryRow(ctx, "SELECT COUNT(*) FROM transactions"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting transactions: %w", err)
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	limit := filter.Limit
	if limit < 1 {
		limit = 20
	}
	offset := (page - 1) * limit

	query := "SELECT " + txColumns + " FROM transactions" + where +
		" ORDER BY transaction_date DESC LIMIT $" + strconv.Itoa(len(args)+1) +
		" OFFSET $" + strconv.Itoa(len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db(ctx).Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("querying transactions: %w", err)
	}
	defer rows.Close()

	txs, err := collectRows(rows)
	if err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}

func (r *TransactionRepository) Update(ctx context.Context, tx *transaction.Transaction) error {
	query := `
		UPDATE transactions
		SET from_account_id = $2, to_account_id = $3, category_id = $4, amount = $5, note = $6, description = $7, transaction_date = $8
		WHERE id = $1
	`

	cmdTag, err := r.db(ctx).Exec(
		ctx,
		query,
		tx.ID,
		nullIfEmpty(tx.FromAccountID),
		nullIfEmpty(tx.ToAccountID),
		nullIfEmpty(tx.CategoryID),
		tx.Amount,
		nullIfEmpty(tx.Note),
		nullIfEmpty(tx.Description),
		tx.Date,
	)
	if err != nil {
		return fmt.Errorf("updating transaction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return transaction.ErrTransactionNotFound
	}

	return nil
}

func (r *TransactionRepository) Delete(ctx context.Context, id string) error {
	cmdTag, err := r.db(ctx).Exec(ctx, `DELETE FROM transactions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting transaction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return transaction.ErrTransactionNotFound
	}

	return nil
}

func (r *TransactionRepository) Search(ctx context.Context, userID, q string, limit int) ([]*transaction.Transaction, error) {
	if limit < 1 {
		limit = 20
	}

	query := `SELECT ` + txColumns + ` FROM transactions
		WHERE user_id = $1 AND (note ILIKE $2 OR description ILIKE $2)
		ORDER BY transaction_date DESC LIMIT $3`

	rows, err := r.db(ctx).Query(ctx, query, userID, "%"+q+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("searching transactions: %w", err)
	}
	defer rows.Close()

	return collectRows(rows)
}

// SpentByCategoryMonth returns the total EXPENSE amount for a user's category in a given
// YYYY-MM month. It backs the budget domain's spent/remaining calculations.
func (r *TransactionRepository) SpentByCategoryMonth(ctx context.Context, userID, categoryID, month string) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 AND category_id = $2 AND type = 'EXPENSE'
		  AND to_char(transaction_date, 'YYYY-MM') = $3
	`
	var total int64
	if err := r.db(ctx).QueryRow(ctx, query, userID, categoryID, month).Scan(&total); err != nil {
		return 0, fmt.Errorf("summing spent for category month: %w", err)
	}
	return total, nil
}

func collectRows(rows pgx.Rows) ([]*transaction.Transaction, error) {
	var txs []*transaction.Transaction
	for rows.Next() {
		tx, err := scanRow(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning transaction row: %w", err)
		}
		txs = append(txs, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction rows: %w", err)
	}

	return txs, nil
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}
