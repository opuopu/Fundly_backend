package config

import (
	"context"
	"errors"

	domainConfig "github.com/Dodul01/fundly/internal/domain/config"
)

type ConfigResponse struct {
	DefaultCurrency string `json:"default_currency"`
	Language        string `json:"language"`
	Theme           string `json:"theme"`
	UpdatedAt       string `json:"updated_at"`
}

func toResponse(c *domainConfig.Config) ConfigResponse {
	return ConfigResponse{
		DefaultCurrency: c.DefaultCurrency,
		Language:        c.Language,
		Theme:           c.Theme,
		UpdatedAt:       c.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// GetConfigUseCase returns the user's configuration, falling back to system defaults when
// none has been saved yet.
type GetConfigUseCase struct {
	repo domainConfig.Repository
}

func NewGetConfigUseCase(repo domainConfig.Repository) *GetConfigUseCase {
	return &GetConfigUseCase{repo: repo}
}

func (uc *GetConfigUseCase) Execute(ctx context.Context, userID string) (*ConfigResponse, error) {
	c, err := uc.repo.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, domainConfig.ErrConfigNotFound) {
			res := toResponse(domainConfig.Default(userID))
			return &res, nil
		}
		return nil, err
	}
	res := toResponse(c)
	return &res, nil
}

type UpdateConfigInput struct {
	UserID          string  `json:"-"`
	DefaultCurrency *string `json:"default_currency"`
	Language        *string `json:"language"`
	Theme           *string `json:"theme"`
}

// UpdateConfigUseCase updates (or creates) the user's configuration.
type UpdateConfigUseCase struct {
	repo domainConfig.Repository
}

func NewUpdateConfigUseCase(repo domainConfig.Repository) *UpdateConfigUseCase {
	return &UpdateConfigUseCase{repo: repo}
}

func (uc *UpdateConfigUseCase) Execute(ctx context.Context, input UpdateConfigInput) (*ConfigResponse, error) {
	c, err := uc.repo.Get(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, domainConfig.ErrConfigNotFound) {
			c = domainConfig.Default(input.UserID)
		} else {
			return nil, err
		}
	}

	if err := c.Update(input.DefaultCurrency, input.Language, input.Theme); err != nil {
		return nil, err
	}

	if err := uc.repo.Upsert(ctx, c); err != nil {
		return nil, err
	}

	res := toResponse(c)
	return &res, nil
}
