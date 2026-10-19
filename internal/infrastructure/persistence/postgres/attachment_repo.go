package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/attachment"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const attachmentColumns = `id, user_id, transaction_id, file_name, file_url, content_type, file_size, created_at`

type AttachmentRepository struct {
	pool *pgxpool.Pool
}

func NewAttachmentRepository(pool *pgxpool.Pool) *AttachmentRepository {
	return &AttachmentRepository{pool: pool}
}

func scanAttachment(row pgx.Row) (*attachment.Attachment, error) {
	var (
		a      attachment.Attachment
		userID *string
		txID   *string
	)
	if err := row.Scan(
		&a.ID,
		&userID,
		&txID,
		&a.FileName,
		&a.FileURL,
		&a.ContentType,
		&a.FileSize,
		&a.CreatedAt,
	); err != nil {
		return nil, err
	}
	if userID != nil {
		a.UserID = *userID
	}
	if txID != nil {
		a.TransactionID = *txID
	}
	return &a, nil
}

func (r *AttachmentRepository) Create(ctx context.Context, a *attachment.Attachment) error {
	query := `
		INSERT INTO attachments (id, user_id, transaction_id, file_name, file_url, content_type, file_size, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		a.ID, nullIfEmpty(a.UserID), nullIfEmpty(a.TransactionID),
		a.FileName, a.FileURL, a.ContentType, a.FileSize, a.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting attachment: %w", err)
	}
	return nil
}

func (r *AttachmentRepository) GetByID(ctx context.Context, id string) (*attachment.Attachment, error) {
	query := `SELECT ` + attachmentColumns + ` FROM attachments WHERE id = $1`
	a, err := scanAttachment(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, attachment.ErrAttachmentNotFound
		}
		return nil, fmt.Errorf("querying attachment by id: %w", err)
	}
	return a, nil
}

func (r *AttachmentRepository) Delete(ctx context.Context, id string) error {
	cmdTag, err := r.pool.Exec(ctx, `DELETE FROM attachments WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting attachment: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return attachment.ErrAttachmentNotFound
	}
	return nil
}

func (r *AttachmentRepository) ListByTransactionID(ctx context.Context, transactionID string) ([]*attachment.Attachment, error) {
	query := `SELECT ` + attachmentColumns + ` FROM attachments WHERE transaction_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("querying attachments by transaction: %w", err)
	}
	defer rows.Close()

	var out []*attachment.Attachment
	for rows.Next() {
		a, err := scanAttachment(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning attachment row: %w", err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating attachment rows: %w", err)
	}
	return out, nil
}

func (r *AttachmentRepository) CountByTransactionID(ctx context.Context, transactionID string) (int, error) {
	var count int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM attachments WHERE transaction_id = $1`, transactionID).Scan(&count); err != nil {
		return 0, fmt.Errorf("counting attachments: %w", err)
	}
	return count, nil
}

func (r *AttachmentRepository) LinkToTransaction(ctx context.Context, userID, transactionID string, attachmentIDs []string) error {
	if len(attachmentIDs) == 0 {
		return nil
	}
	query := `UPDATE attachments SET transaction_id = $1 WHERE user_id = $2 AND id = ANY($3)`
	if _, err := r.pool.Exec(ctx, query, transactionID, userID, attachmentIDs); err != nil {
		return fmt.Errorf("linking attachments to transaction: %w", err)
	}
	return nil
}
