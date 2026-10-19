package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	statapp "github.com/Dodul01/fundly/internal/application/stat"
	domainStat "github.com/Dodul01/fundly/internal/domain/stat"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type statRepoStub struct {
	breakdown []domainStat.CategoryAmount
	totals    domainStat.Totals
}

func (m statRepoStub) BreakdownByCategory(_ context.Context, _, _, _, _ string) ([]domainStat.CategoryAmount, error) {
	return m.breakdown, nil
}
func (m statRepoStub) MonthlyTotals(_ context.Context, _, _ string) (domainStat.Totals, error) {
	return m.totals, nil
}

// budgetRepoStub (from budget_handler_test.go) satisfies domainBudget.Repository and is
// reused here for the overview use case.

func setupStatHandler(stats statRepoStub) *infrahttp.StatHandler {
	cats := catRepoStub{}
	budgets := &budgetRepoStub{budgets: nil}
	return infrahttp.NewStatHandler(
		statapp.NewBreakdownUseCase(stats, cats, nil),
		statapp.NewOverviewUseCase(stats, budgets, nil),
	)
}

func TestStatHandler_Breakdown(t *testing.T) {
	handler := setupStatHandler(statRepoStub{breakdown: []domainStat.CategoryAmount{
		{CategoryID: "cat_salary", Amount: 6000},
		{CategoryID: "cat_kpi", Amount: 4000},
	}})

	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/stats/breakdown?type=income&month=2026-08", nil), "usr_1")
	rec := httptest.NewRecorder()

	handler.Breakdown(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Data statapp.BreakdownResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if resp.Data.TotalAmount != 10000 || resp.Data.Breakdown[0].Percentage != 60.0 {
		t.Errorf("unexpected breakdown: %+v", resp.Data)
	}
}

func TestStatHandler_Breakdown_InvalidType(t *testing.T) {
	handler := setupStatHandler(statRepoStub{})
	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/stats/breakdown?type=bogus", nil), "usr_1")
	rec := httptest.NewRecorder()

	handler.Breakdown(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestStatHandler_Overview(t *testing.T) {
	handler := setupStatHandler(statRepoStub{totals: domainStat.Totals{TotalIncome: 50000, TotalExpense: 20000}})
	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/stats/overview?month=2026-09", nil), "usr_1")
	rec := httptest.NewRecorder()

	handler.Overview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Data statapp.OverviewResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Data.NetBalance != 30000 {
		t.Errorf("expected net 30000, got %d", resp.Data.NetBalance)
	}
}

func TestStatHandler_Unauthorized(t *testing.T) {
	handler := setupStatHandler(statRepoStub{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats/overview", nil)
	rec := httptest.NewRecorder()

	handler.Overview(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
