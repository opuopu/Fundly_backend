package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ConfigRepository struct {
	pool *pgxpool.Pool
}

func NewConfigRepository(pool *pgxpool.Pool) *ConfigRepository {
	return &ConfigRepository{pool: pool}
}

func (r *ConfigRepository) Get(ctx context.Context, userID string) (*config.Config, error) {
	query := `
		SELECT user_id, default_currency, language, theme, updated_at
		FROM user_configs
		WHERE user_id = $1
	`
	var c config.Config
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&c.UserID,
		&c.DefaultCurrency,
		&c.Language,
		&c.Theme,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, config.ErrConfigNotFound
		}
		return nil, fmt.Errorf("querying user config: %w", err)
	}
	return &c, nil
}

func (r *ConfigRepository) Upsert(ctx context.Context, c *config.Config) error {
	query := `
		INSERT INTO user_configs (user_id, default_currency, language, theme, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE
		SET default_currency = EXCLUDED.default_currency,
		    language = EXCLUDED.language,
		    theme = EXCLUDED.theme,
		    updated_at = EXCLUDED.updated_at
	`
	_, err := r.pool.Exec(ctx, query, c.UserID, c.DefaultCurrency, c.Language, c.Theme, c.UpdatedAt)
	if err != nil {
		return fmt.Errorf("upserting user config: %w", err)
	}
	return nil
}
