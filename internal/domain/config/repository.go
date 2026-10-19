package config

import "context"

type Repository interface {
	// Get returns the user's configuration or ErrConfigNotFound if none exists yet.
	Get(ctx context.Context, userID string) (*Config, error)
	// Upsert inserts or updates the user's configuration.
	Upsert(ctx context.Context, c *Config) error
}
