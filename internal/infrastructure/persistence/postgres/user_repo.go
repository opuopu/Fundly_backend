package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dodul01/fundly/internal/domain/user"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	query := `
		SELECT u.id, u.email, u.full_name, u.password_hash, u.avatar_url, u.otp, u.otp_expires_at, u.created_at, u.updated_at,
		       COALESCE(c.default_currency, 'USD')
		FROM users u
		LEFT JOIN user_configs c ON u.id = c.user_id
		WHERE u.id = $1
	`

	var u user.User
	var passwordHash *string
	var avatarURL *string
	var otp *string
	var otpExpiresAt *time.Time
	var defaultCurrency string

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Email,
		&u.FullName,
		&passwordHash,
		&avatarURL,
		&otp,
		&otpExpiresAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&defaultCurrency,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, fmt.Errorf("querying user by id: %w", err)
	}

	if passwordHash != nil {
		u.PasswordHash = *passwordHash
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	if otp != nil {
		u.OTP = *otp
	}
	if otpExpiresAt != nil {
		u.OTPExpiresAt = *otpExpiresAt
	}
	u.DefaultCurrency = defaultCurrency

	return &u, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `
		SELECT u.id, u.email, u.full_name, u.password_hash, u.avatar_url, u.otp, u.otp_expires_at, u.created_at, u.updated_at,
		       COALESCE(c.default_currency, 'USD')
		FROM users u
		LEFT JOIN user_configs c ON u.id = c.user_id
		WHERE u.email = $1
	`

	var u user.User
	var passwordHash *string
	var avatarURL *string
	var otp *string
	var otpExpiresAt *time.Time
	var defaultCurrency string

	err := r.pool.QueryRow(ctx, query, email).Scan(
		&u.ID,
		&u.Email,
		&u.FullName,
		&passwordHash,
		&avatarURL,
		&otp,
		&otpExpiresAt,
		&u.CreatedAt,
		&u.UpdatedAt,
		&defaultCurrency,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Return nil when not found, matching auth/register flow
		}
		return nil, fmt.Errorf("querying user by email: %w", err)
	}

	if passwordHash != nil {
		u.PasswordHash = *passwordHash
	}
	if avatarURL != nil {
		u.AvatarURL = *avatarURL
	}
	if otp != nil {
		u.OTP = *otp
	}
	if otpExpiresAt != nil {
		u.OTPExpiresAt = *otpExpiresAt
	}
	u.DefaultCurrency = defaultCurrency

	return &u, nil
}

func (r *UserRepository) Save(ctx context.Context, u *user.User) error {
	query := `
		INSERT INTO users (id, email, full_name, password_hash, avatar_url, otp, otp_expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE
		SET email = EXCLUDED.email,
		    full_name = EXCLUDED.full_name,
		    password_hash = EXCLUDED.password_hash,
		    avatar_url = EXCLUDED.avatar_url,
		    otp = EXCLUDED.otp,
		    otp_expires_at = EXCLUDED.otp_expires_at,
		    updated_at = EXCLUDED.updated_at
	`

	var passVal *string
	if u.PasswordHash != "" {
		passVal = &u.PasswordHash
	}

	var avatarVal *string
	if u.AvatarURL != "" {
		avatarVal = &u.AvatarURL
	}

	var otpVal *string
	if u.OTP != "" {
		otpVal = &u.OTP
	}

	var otpExpVal *time.Time
	if !u.OTPExpiresAt.IsZero() {
		otpExpVal = &u.OTPExpiresAt
	}

	_, err := r.pool.Exec(
		ctx,
		query,
		u.ID,
		u.Email,
		u.FullName,
		passVal,
		avatarVal,
		otpVal,
		otpExpVal,
		u.CreatedAt,
		u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("saving user: %w", err)
	}

	// Ensure user_configs row exists
	currency := u.DefaultCurrency
	if currency == "" {
		currency = "USD"
	}
	configQuery := `
		INSERT INTO user_configs (user_id, default_currency, language, theme, updated_at)
		VALUES ($1, $2, 'en', 'dark', $3)
		ON CONFLICT (user_id) DO NOTHING
	`
	_, _ = r.pool.Exec(ctx, configQuery, u.ID, currency, u.UpdatedAt)

	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	return r.Save(ctx, u)
}
