package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configapp "github.com/Dodul01/fundly/internal/application/config"
	domainConfig "github.com/Dodul01/fundly/internal/domain/config"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type configRepoStub struct {
	configs map[string]*domainConfig.Config
}

func (m *configRepoStub) Get(_ context.Context, userID string) (*domainConfig.Config, error) {
	c, ok := m.configs[userID]
	if !ok {
		return nil, domainConfig.ErrConfigNotFound
	}
	return c, nil
}
func (m *configRepoStub) Upsert(_ context.Context, c *domainConfig.Config) error {
	m.configs[c.UserID] = c
	return nil
}

func setupConfigHandler() (*infrahttp.ConfigHandler, *configRepoStub) {
	repo := &configRepoStub{configs: make(map[string]*domainConfig.Config)}
	return infrahttp.NewConfigHandler(
		configapp.NewGetConfigUseCase(repo),
		configapp.NewUpdateConfigUseCase(repo),
	), repo
}

func TestConfigHandler_GetDefaults(t *testing.T) {
	handler, _ := setupConfigHandler()
	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/configuration", nil), "usr_1")
	rec := httptest.NewRecorder()

	handler.HandleConfiguration(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Data configapp.ConfigResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.DefaultCurrency != "USD" {
		t.Errorf("expected USD default, got %s", resp.Data.DefaultCurrency)
	}
}

func TestConfigHandler_Update(t *testing.T) {
	handler, repo := setupConfigHandler()
	req := authed(httptest.NewRequest(http.MethodPut, "/api/v1/configuration", strings.NewReader(`{"default_currency":"GBP"}`)), "usr_1")
	rec := httptest.NewRecorder()

	handler.HandleConfiguration(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	if repo.configs["usr_1"].DefaultCurrency != "GBP" {
		t.Errorf("expected GBP persisted, got %s", repo.configs["usr_1"].DefaultCurrency)
	}
}

func TestConfigHandler_Update_InvalidCurrency(t *testing.T) {
	handler, _ := setupConfigHandler()
	req := authed(httptest.NewRequest(http.MethodPut, "/api/v1/configuration", strings.NewReader(`{"default_currency":"dollar"}`)), "usr_1")
	rec := httptest.NewRecorder()

	handler.HandleConfiguration(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestConfigHandler_Unauthorized(t *testing.T) {
	handler, _ := setupConfigHandler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/configuration", nil)
	rec := httptest.NewRecorder()

	handler.HandleConfiguration(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
