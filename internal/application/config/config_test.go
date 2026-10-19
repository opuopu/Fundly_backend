package config_test

import (
	"context"
	"errors"
	"testing"

	configapp "github.com/Dodul01/fundly/internal/application/config"
	domainConfig "github.com/Dodul01/fundly/internal/domain/config"
)

type mockConfigRepo struct {
	configs map[string]*domainConfig.Config
}

func newMockRepo() *mockConfigRepo {
	return &mockConfigRepo{configs: make(map[string]*domainConfig.Config)}
}

func (m *mockConfigRepo) Get(_ context.Context, userID string) (*domainConfig.Config, error) {
	c, ok := m.configs[userID]
	if !ok {
		return nil, domainConfig.ErrConfigNotFound
	}
	return c, nil
}

func (m *mockConfigRepo) Upsert(_ context.Context, c *domainConfig.Config) error {
	m.configs[c.UserID] = c
	return nil
}

func TestGetConfig_DefaultsWhenMissing(t *testing.T) {
	uc := configapp.NewGetConfigUseCase(newMockRepo())
	res, err := uc.Execute(context.Background(), "usr_1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DefaultCurrency != "USD" || res.Language != "en" || res.Theme != "dark" {
		t.Errorf("expected system defaults, got %+v", res)
	}
}

func TestUpdateConfig_CreatesAndUpdates(t *testing.T) {
	repo := newMockRepo()
	uc := configapp.NewUpdateConfigUseCase(repo)

	cur := "eur"
	res, err := uc.Execute(context.Background(), configapp.UpdateConfigInput{
		UserID:          "usr_1",
		DefaultCurrency: &cur,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.DefaultCurrency != "EUR" {
		t.Errorf("expected EUR (upper-cased), got %s", res.DefaultCurrency)
	}
	if _, ok := repo.configs["usr_1"]; !ok {
		t.Error("expected config persisted")
	}
}

func TestUpdateConfig_InvalidCurrency(t *testing.T) {
	uc := configapp.NewUpdateConfigUseCase(newMockRepo())
	bad := "US"
	_, err := uc.Execute(context.Background(), configapp.UpdateConfigInput{UserID: "usr_1", DefaultCurrency: &bad})
	if !errors.Is(err, domainConfig.ErrInvalidCurrency) {
		t.Fatalf("expected ErrInvalidCurrency, got %v", err)
	}
}

func TestUpdateConfig_NoFields(t *testing.T) {
	uc := configapp.NewUpdateConfigUseCase(newMockRepo())
	_, err := uc.Execute(context.Background(), configapp.UpdateConfigInput{UserID: "usr_1"})
	if !errors.Is(err, domainConfig.ErrNoFieldsToUpdate) {
		t.Fatalf("expected ErrNoFieldsToUpdate, got %v", err)
	}
}
