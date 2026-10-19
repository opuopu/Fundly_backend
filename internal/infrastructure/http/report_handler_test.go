package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	reportapp "github.com/Dodul01/fundly/internal/application/report"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

func setupReportHandler(stats statRepoStub) *infrahttp.ReportHandler {
	return infrahttp.NewReportHandler(reportapp.NewCompareUseCase(stats, catRepoStub{}, nil))
}

func TestReportHandler_Compare(t *testing.T) {
	// statRepoStub ignores month, returning the same breakdown for both periods, so the
	// difference nets to zero — enough to exercise the handler wiring and envelope.
	handler := setupReportHandler(statRepoStub{breakdown: nil})

	body := `{"type":"INCOME","primary_period":{"month":"2026-08"},"comparison_period":{"month":"2026-07"},"category_ids":["cat_salary"]}`
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/reports/compare", strings.NewReader(body)), "usr_1")
	rec := httptest.NewRecorder()

	handler.Compare(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data reportapp.CompareResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if resp.Data.PrimaryPeriod != "2026-08" || resp.Data.ComparisonPeriod != "2026-07" {
		t.Errorf("unexpected periods: %+v", resp.Data)
	}
}

func TestReportHandler_Compare_InvalidType(t *testing.T) {
	handler := setupReportHandler(statRepoStub{})
	body := `{"type":"bogus","primary_period":{"month":"2026-08"}}`
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/reports/compare", strings.NewReader(body)), "usr_1")
	rec := httptest.NewRecorder()

	handler.Compare(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestReportHandler_Unauthorized(t *testing.T) {
	handler := setupReportHandler(statRepoStub{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports/compare", strings.NewReader(`{"type":"INCOME"}`))
	rec := httptest.NewRecorder()

	handler.Compare(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
