package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	termsapp "github.com/Dodul01/fundly/internal/application/terms"
	domainTerms "github.com/Dodul01/fundly/internal/domain/terms"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type termsRepoStub struct {
	latest *domainTerms.TermsAndConditions
	byVer  map[string]*domainTerms.TermsAndConditions
	acks   int
}

func (m *termsRepoStub) GetLatest(_ context.Context) (*domainTerms.TermsAndConditions, error) {
	if m.latest == nil {
		return nil, domainTerms.ErrTermsNotFound
	}
	return m.latest, nil
}
func (m *termsRepoStub) GetByVersion(_ context.Context, version string) (*domainTerms.TermsAndConditions, error) {
	t, ok := m.byVer[version]
	if !ok {
		return nil, domainTerms.ErrTermsNotFound
	}
	return t, nil
}
func (m *termsRepoStub) SaveAcknowledgment(_ context.Context, _ *domainTerms.Acknowledgment) error {
	m.acks++
	return nil
}

func setupTermsHandler(repo *termsRepoStub) *infrahttp.TermsHandler {
	return infrahttp.NewTermsHandler(
		termsapp.NewGetLatestUseCase(repo),
		termsapp.NewAcknowledgeUseCase(repo),
	)
}

func TestTermsHandler_GetLatest_Public(t *testing.T) {
	repo := &termsRepoStub{latest: &domainTerms.TermsAndConditions{
		Version:   "2.1",
		Content:   "Only clause.",
		Effective: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}
	handler := setupTermsHandler(repo)

	// No auth context — endpoint is public.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/terms/latest", nil)
	rec := httptest.NewRecorder()

	handler.GetLatest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data termsapp.LatestTermsResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.Version != "2.1" || len(resp.Data.Clauses) != 1 {
		t.Errorf("unexpected terms response: %+v", resp.Data)
	}
}

func TestTermsHandler_Acknowledge(t *testing.T) {
	repo := &termsRepoStub{byVer: map[string]*domainTerms.TermsAndConditions{"2.1": {Version: "2.1"}}}
	handler := setupTermsHandler(repo)

	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/terms/acknowledge", strings.NewReader(`{"version":"2.1"}`)), "usr_1")
	rec := httptest.NewRecorder()

	handler.Acknowledge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	if repo.acks != 1 {
		t.Errorf("expected 1 acknowledgment, got %d", repo.acks)
	}
}

func TestTermsHandler_Acknowledge_Unauthorized(t *testing.T) {
	handler := setupTermsHandler(&termsRepoStub{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/terms/acknowledge", strings.NewReader(`{"version":"2.1"}`))
	rec := httptest.NewRecorder()

	handler.Acknowledge(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestTermsHandler_Acknowledge_UnknownVersion(t *testing.T) {
	handler := setupTermsHandler(&termsRepoStub{byVer: map[string]*domainTerms.TermsAndConditions{}})
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/terms/acknowledge", strings.NewReader(`{"version":"9.9"}`)), "usr_1")
	rec := httptest.NewRecorder()

	handler.Acknowledge(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
