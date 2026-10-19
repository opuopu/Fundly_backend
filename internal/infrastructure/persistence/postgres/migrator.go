package postgres

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Dodul01/fundly/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyMigrations executes all embedded "*.up.sql" migration scripts in filename order.
// Scripts are written to be idempotent (CREATE TABLE IF NOT EXISTS, ADD COLUMN IF NOT
// EXISTS, ON CONFLICT DO NOTHING), so re-running them on every boot is safe.
func ApplyMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return fmt.Errorf("reading embedded migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		sqlBytes, err := migrations.FS.ReadFile(name)
		if err != nil {
			return fmt.Errorf("reading embedded migration %s: %w", name, err)
		}
		if _, err := pool.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("executing migration %s: %w", name, err)
		}
	}

	return nil
}
