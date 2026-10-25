package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/privacypolicy"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrivacyPolicyRepository struct {
	pool *pgxpool.Pool
}

func NewPrivacyPolicyRepository(pool *pgxpool.Pool) *PrivacyPolicyRepository {
	return &PrivacyPolicyRepository{pool: pool}
}

func (r *PrivacyPolicyRepository) GetLatest(ctx context.Context) (*privacypolicy.PrivacyPolicy, error) {
	query := `
		SELECT id, version, content, effective_date
		FROM privacy_policies
		ORDER BY effective_date DESC, created_at DESC
		LIMIT 1
	`
	p, err := scanPrivacyPolicy(r.pool.QueryRow(ctx, query))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, privacypolicy.ErrPolicyNotFound
		}
		return nil, fmt.Errorf("querying latest privacy policy: %w", err)
	}
	return p, nil
}

func (r *PrivacyPolicyRepository) GetByVersion(ctx context.Context, version string) (*privacypolicy.PrivacyPolicy, error) {
	query := `
		SELECT id, version, content, effective_date
		FROM privacy_policies
		WHERE version = $1
	`
	p, err := scanPrivacyPolicy(r.pool.QueryRow(ctx, query, version))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, privacypolicy.ErrPolicyNotFound
		}
		return nil, fmt.Errorf("querying privacy policy by version: %w", err)
	}
	return p, nil
}

func (r *PrivacyPolicyRepository) SaveAcknowledgment(ctx context.Context, a *privacypolicy.Acknowledgment) error {
	query := `
		INSERT INTO user_privacy_policy_acknowledgments (user_id, version, acknowledged_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, version) DO UPDATE SET acknowledged_at = EXCLUDED.acknowledged_at
	`
	if _, err := r.pool.Exec(ctx, query, a.UserID, a.Version, a.AcknowledgedAt); err != nil {
		return fmt.Errorf("saving privacy policy acknowledgment: %w", err)
	}
	return nil
}

func scanPrivacyPolicy(row pgx.Row) (*privacypolicy.PrivacyPolicy, error) {
	var p privacypolicy.PrivacyPolicy
	if err := row.Scan(&p.ID, &p.Version, &p.Content, &p.Effective); err != nil {
		return nil, err
	}
	return &p, nil
}
