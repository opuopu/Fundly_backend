package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	userapp "github.com/Dodul01/fundly/internal/application/user"
	domainUser "github.com/Dodul01/fundly/internal/domain/user"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
	"github.com/Dodul01/fundly/internal/infrastructure/security"
)

type mockStorageForHandler struct{}

func (m *mockStorageForHandler) SaveAvatar(ctx context.Context, userID, filename, contentType string, reader io.Reader) (string, error) {
	return "/uploads/avatars/" + filename, nil
}

type mockUserRepoForHandler struct {
	usersByID    map[string]*domainUser.User
	usersByEmail map[string]*domainUser.User
	err          error
}

func (m *mockUserRepoForHandler) GetByID(ctx context.Context, id string) (*domainUser.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	u, ok := m.usersByID[id]
	if !ok {
		return nil, domainUser.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepoForHandler) FindByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.usersByEmail != nil {
		return m.usersByEmail[email], nil
	}
	return nil, nil
}

func (m *mockUserRepoForHandler) Save(ctx context.Context, user *domainUser.User) error {
	if m.usersByID != nil {
		m.usersByID[user.ID] = user
	}
	if m.usersByEmail != nil {
		m.usersByEmail[user.Email] = user
	}
	return nil
}

func (m *mockUserRepoForHandler) Update(ctx context.Context, user *domainUser.User) error {
	return m.Save(ctx, user)
}

func setupTestUserHandler(repo *mockUserRepoForHandler) *infrahttp.UserHandler {
	getProfileUC := userapp.NewGetProfileUseCase(repo)
	updateProfileUC := userapp.NewUpdateProfileUseCase(repo)
	hasher := security.NewBcryptHasher()
	changePasswordUC := userapp.NewChangePasswordUseCase(repo, hasher)
	uploadAvatarUC := userapp.NewUploadAvatarUseCase(repo, &mockStorageForHandler{})
	return infrahttp.NewUserHandler(getProfileUC, updateProfileUC, changePasswordUC, uploadAvatarUC)
}

func TestUserHandler_GetProfile_Success(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_123", "alice@example.com", "Alice Haye", "hashed_pwd")
	u.AvatarURL = "https://assets.fundly.com/avatars/alice.png"
	u.DefaultCurrency = "USD"
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                   `json:"success"`
		Data    userapp.ProfileResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success to be true")
	}
	if resp.Data.ID != "usr_123" {
		t.Errorf("expected ID 'usr_123', got '%s'", resp.Data.ID)
	}
	if resp.Data.Email != "alice@example.com" {
		t.Errorf("expected email 'alice@example.com', got '%s'", resp.Data.Email)
	}
	if resp.Data.FullName != "Alice Haye" {
		t.Errorf("expected full_name 'Alice Haye', got '%s'", resp.Data.FullName)
	}
	if resp.Data.AvatarURL != "https://assets.fundly.com/avatars/alice.png" {
		t.Errorf("expected avatar_url, got '%s'", resp.Data.AvatarURL)
	}
	if resp.Data.DefaultCurrency != "USD" {
		t.Errorf("expected default_currency 'USD', got '%s'", resp.Data.DefaultCurrency)
	}
}

func TestUserHandler_GetProfile_MissingContext(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestUserHandler_GetProfile_NotFound(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_unknown")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}
}

func TestUserHandler_UpdateProfile_Success(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_123", "alice@example.com", "Alice Haye", "hashed_pwd")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	handler := setupTestUserHandler(repo)

	body := []byte(`{"full_name":"Thom Haye","email":"thomhaye@example.com"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    userapp.ProfileResponse `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success to be true")
	}
	if resp.Data.FullName != "Thom Haye" {
		t.Errorf("expected full_name 'Thom Haye', got '%s'", resp.Data.FullName)
	}
	if resp.Data.Email != "thomhaye@example.com" {
		t.Errorf("expected email 'thomhaye@example.com', got '%s'", resp.Data.Email)
	}
	if resp.Message != "Profile updated successfully" {
		t.Errorf("expected confirmation message, got '%s'", resp.Message)
	}
}

func TestUserHandler_UpdateProfile_EmailConflict(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	u1, _ := domainUser.New("usr_1", "user1@example.com", "User One", "hashed_pwd")
	u2, _ := domainUser.New("usr_2", "taken@example.com", "User Two", "hashed_pwd")
	repo.usersByID[u1.ID] = u1
	repo.usersByEmail[u1.Email] = u1
	repo.usersByID[u2.ID] = u2
	repo.usersByEmail[u2.Email] = u2

	handler := setupTestUserHandler(repo)

	body := []byte(`{"email":"taken@example.com"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", rec.Code)
	}
}

func TestUserHandler_UpdateProfile_InvalidEmail(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_1", "user1@example.com", "User One", "hashed_pwd")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	handler := setupTestUserHandler(repo)

	body := []byte(`{"email":"not-an-email"}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_UpdateProfile_InvalidJSON(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID: make(map[string]*domainUser.User),
	}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader([]byte("{invalid-json")))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_UpdateProfile_EmptyFields(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_1", "user1@example.com", "User One", "hashed_pwd")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	handler := setupTestUserHandler(repo)

	body := []byte(`{"full_name":"","email":""}`)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_MethodNotAllowed(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/users/me", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.HandleMe(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestUserHandler_WithAuthMiddleware(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_456", "bob@example.com", "Bob Ross", "hashed_pwd")
	repo.usersByID[u.ID] = u
	repo.usersByEmail[u.Email] = u

	handler := setupTestUserHandler(repo)

	jwtSvc := security.NewJWTService("super-secret-key-123", time.Hour, 24*time.Hour)
	blacklist := security.NewMemoryTokenBlacklist()
	authMiddleware := infrahttp.NewAuthMiddleware(jwtSvc, blacklist)

	protectedHandler := authMiddleware.RequireAuth(handler.HandleMe)

	// 1. Missing Authorization header
	reqNoAuth := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	recNoAuth := httptest.NewRecorder()
	protectedHandler(recNoAuth, reqNoAuth)
	if recNoAuth.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing auth header, got %d", recNoAuth.Code)
	}

	// 2. Valid token GET
	tokenPair, err := jwtSvc.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	reqValid := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	reqValid.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	recValid := httptest.NewRecorder()
	protectedHandler(recValid, reqValid)
	if recValid.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid token, got %d, body: %s", recValid.Code, recValid.Body.String())
	}

	// 3. Valid token PATCH
	patchBody := []byte(`{"full_name":"Robert Ross"}`)
	reqPatch := httptest.NewRequest(http.MethodPatch, "/api/v1/users/me", bytes.NewReader(patchBody))
	reqPatch.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	recPatch := httptest.NewRecorder()
	protectedHandler(recPatch, reqPatch)
	if recPatch.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid token on PATCH, got %d, body: %s", recPatch.Code, recPatch.Body.String())
	}

	// 4. Blacklisted / Revoked token
	_ = blacklist.RevokeToken(context.Background(), tokenPair.AccessToken)
	reqRevoked := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	reqRevoked.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
	recRevoked := httptest.NewRecorder()
	protectedHandler(recRevoked, reqRevoked)
	if recRevoked.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for revoked token, got %d", recRevoked.Code)
	}
}

func TestUserHandler_ChangePassword_Success(t *testing.T) {
	hasher := security.NewBcryptHasher()
	oldHash, _ := hasher.Hash("OldSecret123!")

	repo := &mockUserRepoForHandler{
		usersByID:    make(map[string]*domainUser.User),
		usersByEmail: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_123", "alice@example.com", "Alice", oldHash)
	repo.usersByID[u.ID] = u

	handler := setupTestUserHandler(repo)

	body := []byte(`{
		"current_password": "OldSecret123!",
		"new_password": "NewSecret456!",
		"confirm_password": "NewSecret456!"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ChangePassword(rec, req)

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
		t.Errorf("expected success to be true")
	}
	if resp.Message != "Password changed successfully" {
		t.Errorf("expected message 'Password changed successfully', got '%s'", resp.Message)
	}
}

func TestUserHandler_ChangePassword_InvalidCurrentPassword(t *testing.T) {
	hasher := security.NewBcryptHasher()
	oldHash, _ := hasher.Hash("RealPassword123!")

	repo := &mockUserRepoForHandler{
		usersByID: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_123", "alice@example.com", "Alice", oldHash)
	repo.usersByID[u.ID] = u

	handler := setupTestUserHandler(repo)

	body := []byte(`{
		"current_password": "WrongPassword123!",
		"new_password": "NewSecret456!",
		"confirm_password": "NewSecret456!"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ChangePassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_ChangePassword_PasswordMismatch(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID: make(map[string]*domainUser.User),
	}
	handler := setupTestUserHandler(repo)

	body := []byte(`{
		"current_password": "OldPassword123!",
		"new_password": "NewSecret456!",
		"confirm_password": "DifferentPassword789!"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ChangePassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_ChangePassword_WeakPassword(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID: make(map[string]*domainUser.User),
	}
	handler := setupTestUserHandler(repo)

	body := []byte(`{
		"current_password": "OldPassword123!",
		"new_password": "short",
		"confirm_password": "short"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader(body))
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ChangePassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_ChangePassword_InvalidMethod(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/password", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.ChangePassword(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestUserHandler_ChangePassword_Unauthenticated(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/users/me/password", bytes.NewReader([]byte("{}")))
	rec := httptest.NewRecorder()
	handler.ChangePassword(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestUserHandler_UploadAvatar_Success(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_123", "alice@example.com", "Alice", "hashed_pwd")
	repo.usersByID[u.ID] = u

	handler := setupTestUserHandler(repo)

	// Create multipart body
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("avatar", "test.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	pngBytes := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	_, _ = part.Write(pngBytes)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.UploadAvatar(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Success bool                       `json:"success"`
		Message string                     `json:"message"`
		Data    userapp.UploadAvatarResult `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !resp.Success {
		t.Errorf("expected success true")
	}
	if resp.Data.AvatarURL == "" {
		t.Errorf("expected non-empty avatar URL")
	}
	if resp.Message != "Avatar uploaded successfully" {
		t.Errorf("expected 'Avatar uploaded successfully', got '%s'", resp.Message)
	}
}

func TestUserHandler_UploadAvatar_MissingFile(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.UploadAvatar(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_UploadAvatar_InvalidFormat(t *testing.T) {
	repo := &mockUserRepoForHandler{
		usersByID: make(map[string]*domainUser.User),
	}
	u, _ := domainUser.New("usr_123", "alice@example.com", "Alice", "hashed_pwd")
	repo.usersByID[u.ID] = u

	handler := setupTestUserHandler(repo)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("avatar", "test.txt")
	_, _ = part.Write([]byte("not an image"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.UploadAvatar(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestUserHandler_UploadAvatar_InvalidMethod(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me/avatar", nil)
	ctx := context.WithValue(req.Context(), infrahttp.UserIDKey, "usr_123")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	handler.UploadAvatar(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", rec.Code)
	}
}

func TestUserHandler_UploadAvatar_Unauthenticated(t *testing.T) {
	repo := &mockUserRepoForHandler{usersByID: make(map[string]*domainUser.User)}
	handler := setupTestUserHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/users/me/avatar", nil)
	rec := httptest.NewRecorder()
	handler.UploadAvatar(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}
