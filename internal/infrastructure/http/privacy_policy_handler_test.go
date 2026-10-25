package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	privacypolicyapp "github.com/Dodul01/fundly/internal/application/privacypolicy"
	domainPrivacyPolicy "github.com/Dodul01/fundly/internal/domain/privacypolicy"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type privacyPolicyRepoStub struct {
	latest *domainPrivacyPolicy.PrivacyPolicy
	byVer  map[string]*domainPrivacyPolicy.PrivacyPolicy
	acks   int
}

func (m *privacyPolicyRepoStub) GetLatest(_ context.Context) (*domainPrivacyPolicy.PrivacyPolicy, error) {
	if m.latest == nil {
		return nil, domainPrivacyPolicy.ErrPolicyNotFound
	}
	return m.latest, nil
}
func (m *privacyPolicyRepoStub) GetByVersion(_ context.Context, version string) (*domainPrivacyPolicy.PrivacyPolicy, error) {
	p, ok := m.byVer[version]
	if !ok {
		return nil, domainPrivacyPolicy.ErrPolicyNotFound
	}
	return p, nil
}
func (m *privacyPolicyRepoStub) SaveAcknowledgment(_ context.Context, _ *domainPrivacyPolicy.Acknowledgment) error {
	m.acks++
	return nil
}

func setupPrivacyPolicyHandler(repo *privacyPolicyRepoStub) *infrahttp.PrivacyPolicyHandler {
	return infrahttp.NewPrivacyPolicyHandler(
		privacypolicyapp.NewGetLatestUseCase(repo),
		privacypolicyapp.NewAcknowledgeUseCase(repo),
	)
}

func TestPrivacyPolicyHandler_GetLatest_Public(t *testing.T) {
	repo := &privacyPolicyRepoStub{latest: &domainPrivacyPolicy.PrivacyPolicy{
		Version:   "1.0",
		Content:   "Only clause.",
		Effective: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}}
	handler := setupPrivacyPolicyHandler(repo)

	// No auth context — endpoint is public.
	req := httptest.NewRequest(http.MethodGet, "/api/v1/privacy-policy/latest", nil)
	rec := httptest.NewRecorder()

	handler.GetLatest(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data privacypolicyapp.LatestPrivacyPolicyResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.Version != "1.0" || len(resp.Data.Clauses) != 1 {
		t.Errorf("unexpected privacy policy response: %+v", resp.Data)
	}
}

func TestPrivacyPolicyHandler_Acknowledge(t *testing.T) {
	repo := &privacyPolicyRepoStub{byVer: map[string]*domainPrivacyPolicy.PrivacyPolicy{"1.0": {Version: "1.0"}}}
	handler := setupPrivacyPolicyHandler(repo)

	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/privacy-policy/acknowledge", strings.NewReader(`{"version":"1.0"}`)), "usr_1")
	rec := httptest.NewRecorder()

	handler.Acknowledge(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	if repo.acks != 1 {
		t.Errorf("expected 1 acknowledgment, got %d", repo.acks)
	}
}

func TestPrivacyPolicyHandler_Acknowledge_Unauthorized(t *testing.T) {
	handler := setupPrivacyPolicyHandler(&privacyPolicyRepoStub{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/privacy-policy/acknowledge", strings.NewReader(`{"version":"1.0"}`))
	rec := httptest.NewRecorder()

	handler.Acknowledge(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPrivacyPolicyHandler_Acknowledge_UnknownVersion(t *testing.T) {
	handler := setupPrivacyPolicyHandler(&privacyPolicyRepoStub{byVer: map[string]*domainPrivacyPolicy.PrivacyPolicy{}})
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/privacy-policy/acknowledge", strings.NewReader(`{"version":"9.9"}`)), "usr_1")
	rec := httptest.NewRecorder()

	handler.Acknowledge(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
