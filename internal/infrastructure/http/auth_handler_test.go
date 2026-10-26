package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
	infrahttp "github.com/Dodul01/fundly/internal/infrastructure/http"
)

type mockUserRepo struct {
	users map[string]*user.User
}

func (m *mockUserRepo) GetByID(_ context.Context, id string) (*user.User, error) {
	if u, ok := m.users[id]; ok {
		return u, nil
	}
	return nil, user.ErrUserNotFound
}

func (m *mockUserRepo) FindByEmail(_ context.Context, email string) (*user.User, error) {
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) Save(_ context.Context, u *user.User) error {
	m.users[u.ID] = u
	return nil
}

func (m *mockUserRepo) Update(_ context.Context, u *user.User) error {
	m.users[u.ID] = u
	return nil
}

type mockHasher struct{}

func (m *mockHasher) Hash(p string) (string, error) {
	return "hashed_" + p, nil
}

func (m *mockHasher) Compare(hashedPassword, password string) error {
	if hashedPassword == "hashed_"+password {
		return nil
	}
	return errors.New("invalid password")
}

type mockOTPGen struct{}

func (m *mockOTPGen) Generate() string {
	return "128333"
}

type mockSocialValidator struct {
	payload *auth.SocialTokenPayload
	err     error
}

func (m *mockSocialValidator) ValidateToken(_ context.Context, provider, idToken string) (*auth.SocialTokenPayload, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.payload, nil
}

type mockRefreshValidator struct {
	userID string
	email  string
	err    error
}

func (m *mockRefreshValidator) ValidateRefreshToken(tokenStr string) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	return m.userID, m.email, nil
}

type mockTokenGen struct{}

func (m *mockTokenGen) GenerateTokenPair(userID, email string) (*auth.TokenResult, error) {
	return &auth.TokenResult{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresIn:    3600,
	}, nil
}

type mockBlacklist struct {
	tokens map[string]bool
}

func (m *mockBlacklist) RevokeToken(_ context.Context, token string) error {
	m.tokens[token] = true
	return nil
}

func (m *mockBlacklist) IsTokenRevoked(_ context.Context, token string) (bool, error) {
	return m.tokens[token], nil
}

func newTestAuthHandler(
	repo *mockUserRepo,
	hasher *mockHasher,
	tokenGen *mockTokenGen,
	socialValidator *mockSocialValidator,
	refreshValidator *mockRefreshValidator,
) *infrahttp.AuthHandler {
	if hasher == nil {
		hasher = &mockHasher{}
	}
	if tokenGen == nil {
		tokenGen = &mockTokenGen{}
	}
	if socialValidator == nil {
		socialValidator = &mockSocialValidator{}
	}
	if refreshValidator == nil {
		refreshValidator = &mockRefreshValidator{userID: "usr-1", email: "user@example.com"}
	}
	otpGen := &mockOTPGen{}
	regUC := auth.NewRegisterUseCase(repo, hasher, otpGen, nil)
	verifyUC := auth.NewVerifyOTPUseCase(repo, tokenGen)
	resendUC := auth.NewResendOTPUseCase(repo, otpGen, nil)
	loginUC := auth.NewLoginUseCase(repo, hasher, tokenGen)
	socialUC := auth.NewSocialLoginUseCase(repo, socialValidator, tokenGen)
	refreshUC := auth.NewRefreshTokenUseCase(repo, refreshValidator, tokenGen)
	forgotPasswordUC := auth.NewForgotPasswordUseCase(repo, otpGen, nil)
	resetPasswordUC := auth.NewResetPasswordUseCase(repo, hasher)
	logoutUC := auth.NewLogoutUseCase(&mockBlacklist{tokens: make(map[string]bool)})
	return infrahttp.NewAuthHandler(regUC, verifyUC, resendUC, loginUC, socialUC, refreshUC, forgotPasswordUC, resetPasswordUC, logoutUC)
}

func TestAuthHandlerRegisterSuccess(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email":            "test@example.com",
		"password":         "Password123!",
		"confirm_password": "Password123!",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerVerifyOTPSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	u.SetOTP("128333", 5*time.Minute)

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email": "user@example.com",
		"otp":   "128333",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/verify-otp", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.VerifyOTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerResendOTPSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email": "user@example.com",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/resend-otp", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.ResendOTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp infrahttp.Response
	_ = json.NewDecoder(rec.Body).Decode(&resp)
	if resp.Message != "New verification code has been dispatched" {
		t.Errorf("expected dispatch message, got %s", resp.Message)
	}
}

func TestAuthHandlerLoginSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_Password123!")
	u.ClearOTP() // Account is verified

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email":    "user@example.com",
		"password": "Password123!",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp infrahttp.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success: true, got false")
	}
}

func TestAuthHandlerLoginInvalidCredentials(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_Password123!")
	u.ClearOTP()

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email":    "user@example.com",
		"password": "WrongPassword",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerSocialLoginSuccess(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	validator := &mockSocialValidator{
		payload: &auth.SocialTokenPayload{
			Provider:  "google",
			Subject:   "google-sub-123",
			Email:     "social@example.com",
			FullName:  "Google User",
			AvatarURL: "https://avatar.url",
		},
	}
	handler := newTestAuthHandler(repo, nil, nil, validator, nil)

	body := map[string]string{
		"provider": "google",
		"id_token": "valid_token",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/social-login", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.SocialLogin(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp infrahttp.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success: true, got false")
	}
}

func TestAuthHandlerSocialLoginUnsupportedProvider(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"provider": "unknown_provider",
		"id_token": "valid_token",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/social-login", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.SocialLogin(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerRefreshSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	u.ClearOTP() // verified
	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	refreshValidator := &mockRefreshValidator{userID: "usr-1", email: "user@example.com"}
	handler := newTestAuthHandler(repo, nil, nil, nil, refreshValidator)

	body := map[string]string{
		"refresh_token": "valid_refresh_token",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.Refresh(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp infrahttp.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success: true, got false")
	}
}

func TestAuthHandlerRefreshInvalidToken(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	refreshValidator := &mockRefreshValidator{err: errors.New("invalid signature")}
	handler := newTestAuthHandler(repo, nil, nil, nil, refreshValidator)

	body := map[string]string{
		"refresh_token": "bad_token",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.Refresh(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerRefreshEmptyToken(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"refresh_token": "",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.Refresh(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerForgotPasswordSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email": "user@example.com",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.ForgotPassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp infrahttp.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success: true, got false")
	}
	if resp.Message != "Password reset code sent to your email" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestAuthHandlerForgotPasswordEmptyEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email": "",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/forgot-password", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.ForgotPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerResetPasswordSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	u.SetOTP("489211", 15*time.Minute)
	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email":        "user@example.com",
		"new_password": "NewSecretPassword123!",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.ResetPassword(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp infrahttp.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success: true, got false")
	}
	if resp.Message != "Password updated successfully" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}

func TestAuthHandlerResetPasswordWeakPassword(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	u.SetOTP("489211", 15*time.Minute)
	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"email":        "user@example.com",
		"new_password": "short",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/reset-password", bytes.NewReader(raw))
	rec := httptest.NewRecorder()

	handler.ResetPassword(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthHandlerLogoutSuccess(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	handler := newTestAuthHandler(repo, nil, nil, nil, nil)

	body := map[string]string{
		"refresh_token": "some_refresh_token",
	}
	raw, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", bytes.NewReader(raw))
	req.Header.Set("Authorization", "Bearer valid_access_token")
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp infrahttp.Response
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Success {
		t.Errorf("expected success: true, got false")
	}
	if resp.Message != "Logged out successfully" {
		t.Errorf("unexpected message: %s", resp.Message)
	}
}



