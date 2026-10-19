package config

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var (
	ErrConfigNotFound   = errors.New("configuration not found")
	ErrInvalidCurrency  = errors.New("default_currency must be a 3-letter ISO 4217 code")
	ErrNoFieldsToUpdate = errors.New("at least one field must be provided to update")
)

const (
	DefaultCurrency = "USD"
	DefaultLanguage = "en"
	DefaultTheme    = "dark"
)

var currencyPattern = regexp.MustCompile(`^[A-Za-z]{3}$`)

type Config struct {
	UserID          string    `json:"-"`
	DefaultCurrency string    `json:"default_currency"`
	Language        string    `json:"language"`
	Theme           string    `json:"theme"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Default returns a configuration populated with the system defaults for a user.
func Default(userID string) *Config {
	return &Config{
		UserID:          userID,
		DefaultCurrency: DefaultCurrency,
		Language:        DefaultLanguage,
		Theme:           DefaultTheme,
		UpdatedAt:       time.Now().UTC(),
	}
}

// Update applies the provided non-nil fields, validating the currency code.
func (c *Config) Update(defaultCurrency, language, theme *string) error {
	if defaultCurrency == nil && language == nil && theme == nil {
		return ErrNoFieldsToUpdate
	}
	if defaultCurrency != nil {
		cur := strings.TrimSpace(*defaultCurrency)
		if !currencyPattern.MatchString(cur) {
			return ErrInvalidCurrency
		}
		c.DefaultCurrency = strings.ToUpper(cur)
	}
	if language != nil {
		if lang := strings.TrimSpace(*language); lang != "" {
			c.Language = lang
		}
	}
	if theme != nil {
		if th := strings.TrimSpace(*theme); th != "" {
			c.Theme = th
		}
	}
	c.UpdatedAt = time.Now().UTC()
	return nil
}
