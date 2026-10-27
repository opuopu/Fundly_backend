package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	budgetapp "github.com/Dodul01/fundly/internal/application/budget"
	domainBudget "github.com/Dodul01/fundly/internal/domain/budget"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type budgetRepoStub struct {
	budgets map[string]*domainBudget.Budget
}

func (m *budgetRepoStub) Create(_ context.Context, b *domainBudget.Budget) error {
	m.budgets[b.ID] = b
	return nil
}
func (m *budgetRepoStub) GetByID(_ context.Context, id string) (*domainBudget.Budget, error) {
	b, ok := m.budgets[id]
	if !ok {
		return nil, domainBudget.ErrBudgetNotFound
	}
	return b, nil
}
func (m *budgetRepoStub) List(_ context.Context, userID, month, categoryID string) ([]*domainBudget.Budget, error) {
	var out []*domainBudget.Budget
	for _, b := range m.budgets {
		if b.UserID == userID {
			out = append(out, b)
		}
	}
	return out, nil
}
func (m *budgetRepoStub) Update(_ context.Context, b *domainBudget.Budget) error {
	m.budgets[b.ID] = b
	return nil
}
func (m *budgetRepoStub) Delete(_ context.Context, id string) error {
	delete(m.budgets, id)
	return nil
}

type spentStub struct{ amount int64 }

func (s spentStub) SpentByCategoryMonth(_ context.Context, _, _, _ string) (int64, error) {
	return s.amount, nil
}

func setupBudgetHandler(spent int64) (*infrahttp.BudgetHandler, *budgetRepoStub) {
	repo := &budgetRepoStub{budgets: make(map[string]*domainBudget.Budget)}
	cats := catRepoStub{}
	sp := spentStub{amount: spent}
	return infrahttp.NewBudgetHandler(
		budgetapp.NewListBudgetsUseCase(repo, cats, sp),
		budgetapp.NewCreateBudgetUseCase(repo, cats, sp),
		budgetapp.NewGetBudgetUseCase(repo, cats, sp),
		budgetapp.NewUpdateBudgetUseCase(repo, cats, sp),
		budgetapp.NewDeleteBudgetUseCase(repo),
	), repo
}

func TestBudgetHandler_Create(t *testing.T) {
	handler, repo := setupBudgetHandler(45480)

	body := `{"month":"2026-09","limit_amount":80000,"category_id":"cat_food","currency":"USD"}`
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/budgets", strings.NewReader(body)), "usr_1")
	rec := httptest.NewRecorder()

	handler.HandleBudgets(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data budgetapp.BudgetResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if resp.Data.Expenditure != 45480 || resp.Data.RemainingAmount != 34520 {
		t.Errorf("unexpected derived figures: %+v", resp.Data)
	}
	if resp.Data.Date != "2026-09-01" {
		t.Errorf("expected date defaulted to 2026-09-01, got %s", resp.Data.Date)
	}
	if len(repo.budgets) != 1 {
		t.Errorf("expected budget persisted, got %d", len(repo.budgets))
	}
}

func TestBudgetHandler_Create_InvalidMonth(t *testing.T) {
	handler, _ := setupBudgetHandler(0)
	body := `{"month":"2026/09","limit_amount":80000,"category_id":"cat_food"}`
	req := authed(httptest.NewRequest(http.MethodPost, "/api/v1/budgets", strings.NewReader(body)), "usr_1")
	rec := httptest.NewRecorder()

	handler.HandleBudgets(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBudgetHandler_Update_OverridesExpenditure(t *testing.T) {
	handler, repo := setupBudgetHandler(45480)
	b, _ := domainBudget.NewBudget("bg_1", "usr_1", "cat_food", "2026-09", 80000, "USD", "", "")
	repo.budgets["bg_1"] = b

	body := `{"expenditure":3000}`
	req := authed(httptest.NewRequest(http.MethodPatch, "/api/v1/budgets/bg_1", strings.NewReader(body)), "usr_1")
	req.SetPathValue("id", "bg_1")
	rec := httptest.NewRecorder()

	handler.HandleBudgetByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data budgetapp.BudgetResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	// The stub's live-transaction total (45480) must be ignored once overridden.
	if resp.Data.Expenditure != 3000 || resp.Data.RemainingAmount != 77000 {
		t.Errorf("unexpected override result: %+v", resp.Data)
	}
}

func TestBudgetHandler_Get_NotFound(t *testing.T) {
	handler, _ := setupBudgetHandler(0)
	req := authed(httptest.NewRequest(http.MethodGet, "/api/v1/budgets/bg_missing", nil), "usr_1")
	req.SetPathValue("id", "bg_missing")
	rec := httptest.NewRecorder()

	handler.HandleBudgetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestBudgetHandler_Unauthorized(t *testing.T) {
	handler, _ := setupBudgetHandler(0)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/budgets", nil)
	rec := httptest.NewRecorder()

	handler.HandleBudgets(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
