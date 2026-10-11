package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/category"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepository struct {
	pool *pgxpool.Pool
}

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

func (r *CategoryRepository) GetAll(ctx context.Context) ([]*category.Category, error) {
	query := `
		SELECT id, name, type, icon, color, created_at
		FROM categories
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("querying all categories: %w", err)
	}
	defer rows.Close()

	var categories []*category.Category
	for rows.Next() {
		var cat category.Category
		var catType string
		if err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&catType,
			&cat.Icon,
			&cat.Color,
			&cat.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning category row: %w", err)
		}
		cat.Type = category.CategoryType(catType)
		categories = append(categories, &cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating category rows: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepository) GetByType(ctx context.Context, catType category.CategoryType) ([]*category.Category, error) {
	query := `
		SELECT id, name, type, icon, color, created_at
		FROM categories
		WHERE type = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, string(catType))
	if err != nil {
		return nil, fmt.Errorf("querying categories by type: %w", err)
	}
	defer rows.Close()

	var categories []*category.Category
	for rows.Next() {
		var cat category.Category
		var cType string
		if err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cType,
			&cat.Icon,
			&cat.Color,
			&cat.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning category row: %w", err)
		}
		cat.Type = category.CategoryType(cType)
		categories = append(categories, &cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating category rows: %w", err)
	}

	return categories, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id string) (*category.Category, error) {
	query := `
		SELECT id, name, type, icon, color, created_at
		FROM categories
		WHERE id = $1
	`

	var cat category.Category
	var catType string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&cat.ID,
		&cat.Name,
		&catType,
		&cat.Icon,
		&cat.Color,
		&cat.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, category.ErrCategoryNotFound
		}
		return nil, fmt.Errorf("querying category by id: %w", err)
	}

	cat.Type = category.CategoryType(catType)
	return &cat, nil
}

func (r *CategoryRepository) Create(ctx context.Context, cat *category.Category) error {
	query := `
		INSERT INTO categories (id, name, type, icon, color, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		cat.ID,
		cat.Name,
		string(cat.Type),
		cat.Icon,
		cat.Color,
		cat.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("inserting category: %w", err)
	}

	return nil
}

func (r *CategoryRepository) Update(ctx context.Context, cat *category.Category) error {
	query := `
		UPDATE categories
		SET name = $2, color = $3, icon = $4
		WHERE id = $1
	`

	cmdTag, err := r.pool.Exec(ctx, query, cat.ID, cat.Name, cat.Color, cat.Icon)
	if err != nil {
		return fmt.Errorf("updating category: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return category.ErrCategoryNotFound
	}

	return nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM categories
		WHERE id = $1
	`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("deleting category: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return category.ErrCategoryNotFound
	}

	return nil
}
