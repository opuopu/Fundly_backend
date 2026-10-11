package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	appcategory "github.com/Dodul01/fundly/internal/application/category"
	domainCategory "github.com/Dodul01/fundly/internal/domain/category"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
	"github.com/Dodul01/fundly/internal/infrastructure/security"
)

type mockCategoryRepoForHandler struct {
	categories map[string]*domainCategory.Category
	err        error
}

func (m *mockCategoryRepoForHandler) GetAll(ctx context.Context) ([]*domainCategory.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainCategory.Category
	for _, c := range m.categories {
		res = append(res, c)
	}
	return res, nil
}

func (m *mockCategoryRepoForHandler) GetByType(ctx context.Context, catType domainCategory.CategoryType) ([]*domainCategory.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	var res []*domainCategory.Category
	for _, c := range m.categories {
		if c.Type == catType {
			res = append(res, c)
		}
	}
	return res, nil
}

func (m *mockCategoryRepoForHandler) GetByID(ctx context.Context, id string) (*domainCategory.Category, error) {
	if m.err != nil {
		return nil, m.err
	}
	c, ok := m.categories[id]
	if !ok {
		return nil, domainCategory.ErrCategoryNotFound
	}
	return c, nil
}

func (m *mockCategoryRepoForHandler) Create(ctx context.Context, cat *domainCategory.Category) error {
	if m.err != nil {
		return m.err
	}
	m.categories[cat.ID] = cat
	return nil
}

func (m *mockCategoryRepoForHandler) Update(ctx context.Context, cat *domainCategory.Category) error {
	if m.err != nil {
		return m.err
	}
	m.categories[cat.ID] = cat
	return nil
}

func (m *mockCategoryRepoForHandler) Delete(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	delete(m.categories, id)
	return nil
}

func setupTestCategoryHandler(repo *mockCategoryRepoForHandler) *infrahttp.CategoryHandler {
	listUC := appcategory.NewListCategoriesUseCase(repo)
	createUC := appcategory.NewCreateCategoryUseCase(repo)
	updateUC := appcategory.NewUpdateCategoryUseCase(repo)
	deleteUC := appcategory.NewDeleteCategoryUseCase(repo)
	return infrahttp.NewCategoryHandler(listUC, createUC, updateUC, deleteUC)
}

func TestCategoryHandler_ListCategories_Success(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_food", "Food", "EXPENSE", "bowl-food", "#FF8A65")
	c2, _ := domainCategory.NewCategory("cat_salary", "Salary", "INCOME", "wallet", "#4CAF50")
	repo.categories[c1.ID] = c1
	repo.categories[c2.ID] = c2
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                         `json:"success"`
		Data    []appcategory.CategoryResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 categories, got %d", len(resp.Data))
	}
}

func TestCategoryHandler_ListCategories_FilterType(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_food", "Food", "EXPENSE", "bowl-food", "#FF8A65")
	c2, _ := domainCategory.NewCategory("cat_salary", "Salary", "INCOME", "wallet", "#4CAF50")
	repo.categories[c1.ID] = c1
	repo.categories[c2.ID] = c2
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories?type=income", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                         `json:"success"`
		Data    []appcategory.CategoryResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 category, got %d", len(resp.Data))
	}
	if resp.Data[0].Type != "income" {
		t.Errorf("expected type income, got %s", resp.Data[0].Type)
	}
}

func TestCategoryHandler_ListCategories_InvalidType(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories?type=invalid_filter", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCategoryHandler_ListCategories_Unauthenticated(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestCategoryHandler_HandleCategories_MethodNotAllowed(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/categories", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestCategoryHandler_WithAuthMiddleware(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	jwtSvc := security.NewJWTService("secret-categories-key", time.Hour, 24*time.Hour)
	blacklist := security.NewMemoryTokenBlacklist()
	authMiddleware := infrahttp.NewAuthMiddleware(jwtSvc, blacklist)

	protectedHandler := authMiddleware.RequireAuth(handler.HandleCategories)

	// Valid token
	tokens, _ := jwtSvc.GenerateTokenPair("usr_1", "user@fundly.com")
	reqValid := httptest.NewRequest(http.MethodGet, "/api/v1/categories", nil)
	reqValid.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	recValid := httptest.NewRecorder()
	protectedHandler(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recValid.Code)
	}
}

func TestCategoryHandler_CreateCategory_Success(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"Gym & Fitness","type":"expense","icon":"dumbbell","color":"#E91E63"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    appcategory.CategoryResponse `json:"data"`
		Message string                     `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if resp.Data.Name != "Gym & Fitness" {
		t.Errorf("expected name 'Gym & Fitness', got %s", resp.Data.Name)
	}
	if resp.Data.Type != "expense" {
		t.Errorf("expected type 'expense', got %s", resp.Data.Type)
	}
	if resp.Data.Icon != "dumbbell" {
		t.Errorf("expected icon 'dumbbell', got %s", resp.Data.Icon)
	}
	if resp.Data.Color != "#E91E63" {
		t.Errorf("expected color '#E91E63', got %s", resp.Data.Color)
	}
}

func TestCategoryHandler_CreateCategory_InvalidJSON(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader("{invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCategoryHandler_CreateCategory_InvalidName(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"","type":"expense"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCategoryHandler_CreateCategory_InvalidType(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"Side Hustle","type":"invalid_type"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCategoryHandler_CreateCategory_Unauthenticated(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"Gym","type":"expense"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	handler.HandleCategories(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestCategoryHandler_UpdateCategory_Success(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[c1.ID] = c1
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"Pharmacy & Medicine","color":"#C2185B","icon":"pills"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/cat_01", strings.NewReader(body))
	req.SetPathValue("id", "cat_01")
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    appcategory.CategoryResponse `json:"data"`
		Message string                     `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if resp.Data.Name != "Pharmacy & Medicine" {
		t.Errorf("expected name 'Pharmacy & Medicine', got %s", resp.Data.Name)
	}
	if resp.Data.Color != "#C2185B" {
		t.Errorf("expected color '#C2185B', got %s", resp.Data.Color)
	}
	if resp.Data.Icon != "pills" {
		t.Errorf("expected icon 'pills', got %s", resp.Data.Icon)
	}
}

func TestCategoryHandler_UpdateCategory_Partial(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[c1.ID] = c1
	handler := setupTestCategoryHandler(repo)

	body := `{"color":"#C2185B"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/cat_01", strings.NewReader(body))
	req.SetPathValue("id", "cat_01")
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var resp struct {
		Success bool                       `json:"success"`
		Data    appcategory.CategoryResponse `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Data.Color != "#C2185B" {
		t.Errorf("expected color '#C2185B', got %s", resp.Data.Color)
	}
	if resp.Data.Name != "Old Name" {
		t.Errorf("expected name preserved 'Old Name', got %s", resp.Data.Name)
	}
}

func TestCategoryHandler_UpdateCategory_NotFound(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/cat_nonexistent", strings.NewReader(body))
	req.SetPathValue("id", "cat_nonexistent")
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestCategoryHandler_UpdateCategory_EmptyName(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[c1.ID] = c1
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"  "}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/cat_01", strings.NewReader(body))
	req.SetPathValue("id", "cat_01")
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCategoryHandler_UpdateCategory_NoFields(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[c1.ID] = c1
	handler := setupTestCategoryHandler(repo)

	body := `{}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/cat_01", strings.NewReader(body))
	req.SetPathValue("id", "cat_01")
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCategoryHandler_UpdateCategory_InvalidJSON(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/cat_01", strings.NewReader("{invalid"))
	req.SetPathValue("id", "cat_01")
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCategoryHandler_UpdateCategory_Unauthenticated(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	body := `{"name":"New Name"}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/categories/cat_01", strings.NewReader(body))
	req.SetPathValue("id", "cat_01")

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestCategoryHandler_DeleteCategory_Success(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	c1, _ := domainCategory.NewCategory("cat_01", "Old Name", "EXPENSE", "old-icon", "#000000")
	repo.categories[c1.ID] = c1
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/cat_01", nil)
	req.SetPathValue("id", "cat_01")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool   `json:"success"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if _, exists := repo.categories["cat_01"]; exists {
		t.Errorf("expected category to be deleted from repo")
	}
}

func TestCategoryHandler_DeleteCategory_NotFound(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/cat_nonexistent", nil)
	req.SetPathValue("id", "cat_nonexistent")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestCategoryHandler_DeleteCategory_Unauthenticated(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/categories/cat_01", nil)
	req.SetPathValue("id", "cat_01")

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestCategoryHandler_HandleCategoryByID_MethodNotAllowed(t *testing.T) {
	repo := &mockCategoryRepoForHandler{categories: make(map[string]*domainCategory.Category)}
	handler := setupTestCategoryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/categories/cat_01", nil)
	req.SetPathValue("id", "cat_01")
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleCategoryByID(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}


