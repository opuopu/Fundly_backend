package postgres

import (
	"context"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/feedback"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FeedbackRepository struct {
	pool *pgxpool.Pool
}

func NewFeedbackRepository(pool *pgxpool.Pool) *FeedbackRepository {
	return &FeedbackRepository{pool: pool}
}

func (r *FeedbackRepository) Create(ctx context.Context, f *feedback.Feedback) error {
	query := `
		INSERT INTO feedbacks (id, user_id, subject, description, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.pool.Exec(ctx, query, f.ID, nullIfEmpty(f.UserID), f.Subject, f.Description, f.CreatedAt)
	if err != nil {
		return fmt.Errorf("inserting feedback: %w", err)
	}
	return nil
}
