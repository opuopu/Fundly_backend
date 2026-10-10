package postgres

import (
	"context"
	"fmt"

	"github.com/Dodul01/fundly/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyMigrations executes the initial schema migration script on PostgreSQL.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migrationSQL, err := migrations.FS.ReadFile("000001_init_schema.up.sql")
	if err != nil {
		return fmt.Errorf("reading embedded migration file: %w", err)
	}

	if _, err := pool.Exec(ctx, string(migrationSQL)); err != nil {
		return fmt.Errorf("executing migration script: %w", err)
	}

	return nil
}
