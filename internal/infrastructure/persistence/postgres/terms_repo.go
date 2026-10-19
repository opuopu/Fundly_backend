package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/terms"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TermsRepository struct {
	pool *pgxpool.Pool
}

func NewTermsRepository(pool *pgxpool.Pool) *TermsRepository {
	return &TermsRepository{pool: pool}
}

func (r *TermsRepository) GetLatest(ctx context.Context) (*terms.TermsAndConditions, error) {
	query := `
		SELECT id, version, content, effective_date
		FROM terms_and_conditions
		ORDER BY effective_date DESC, created_at DESC
		LIMIT 1
	`
	t, err := scanTerms(r.pool.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, terms.ErrTermsNotFound
		}
		return nil, fmt.Errorf("querying latest terms: %w", err)
	}
	return t, nil
}

func (r *TermsRepository) GetByVersion(ctx context.Context, version string) (*terms.TermsAndConditions, error) {
	query := `
		SELECT id, version, content, effective_date
		FROM terms_and_conditions
		WHERE version = $1
	`
	t, err := scanTerms(r.pool.QueryRow(ctx, query, version))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, terms.ErrTermsNotFound
		}
		return nil, fmt.Errorf("querying terms by version: %w", err)
	}
	return t, nil
}

func (r *TermsRepository) SaveAcknowledgment(ctx context.Context, a *terms.Acknowledgment) error {
	query := `
		INSERT INTO user_terms_acknowledgments (user_id, version, acknowledged_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, version) DO UPDATE SET acknowledged_at = EXCLUDED.acknowledged_at
	`
	if _, err := r.pool.Exec(ctx, query, a.UserID, a.Version, a.AcknowledgedAt); err != nil {
		return fmt.Errorf("saving terms acknowledgment: %w", err)
	}
	return nil
}

func scanTerms(row pgx.Row) (*terms.TermsAndConditions, error) {
	var t terms.TermsAndConditions
	if err := row.Scan(&t.ID, &t.Version, &t.Content, &t.Effective); err != nil {
		return nil, err
	}
	return &t, nil
}
