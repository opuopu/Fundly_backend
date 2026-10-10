package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

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

func TestSocialLoginNewUserSuccess(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	validator := &mockSocialValidator{
		payload: &auth.SocialTokenPayload{
			Provider:  "google",
			Subject:   "google-sub-123",
			Email:     "social.user@example.com",
			FullName:  "Google User",
			AvatarURL: "https://lh3.googleusercontent.com/photo.jpg",
		},
	}
	tokenGen := &mockTokenGen{}

	uc := auth.NewSocialLoginUseCase(repo, validator, tokenGen)

	cmd := auth.SocialLoginCommand{
		Provider: "google",
		IDToken:  "valid_google_token",
	}

	result, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.User.Email != "social.user@example.com" {
		t.Errorf("expected email social.user@example.com, got %s", result.User.Email)
	}
	if result.User.FullName != "Google User" {
		t.Errorf("expected full name Google User, got %s", result.User.FullName)
	}
	if result.User.AvatarURL != "https://lh3.googleusercontent.com/photo.jpg" {
		t.Errorf("expected avatar URL set, got %s", result.User.AvatarURL)
	}
	if result.Tokens.AccessToken != "mock_access_token" {
		t.Errorf("expected mock_access_token, got %s", result.Tokens.AccessToken)
	}

	saved, _ := repo.FindByEmail(context.Background(), "social.user@example.com")
	if saved == nil {
		t.Fatalf("expected user to be persisted in repo")
	}
	if !saved.IsVerified() {
		t.Errorf("social user should be verified")
	}
}

func TestSocialLoginExistingUserSuccess(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	existing, _ := user.New("usr-ex-1", "existing@example.com", "Existing Name", "hash")
	existing.SetOTP("123456", 5) // Was unverified
	_ = repo.Save(context.Background(), existing)

	validator := &mockSocialValidator{
		payload: &auth.SocialTokenPayload{
			Provider:  "apple",
			Subject:   "apple-sub-456",
			Email:     "existing@example.com",
			FullName:  "Existing Name",
			AvatarURL: "https://apple.com/avatar.jpg",
		},
	}
	tokenGen := &mockTokenGen{}

	uc := auth.NewSocialLoginUseCase(repo, validator, tokenGen)

	cmd := auth.SocialLoginCommand{
		Provider: "apple",
		IDToken:  "valid_apple_token",
	}

	result, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}

	if result.User.ID != "usr-ex-1" {
		t.Errorf("expected user ID usr-ex-1, got %s", result.User.ID)
	}

	// Should have cleared OTP and updated avatar
	updated, _ := repo.FindByEmail(context.Background(), "existing@example.com")
	if !updated.IsVerified() {
		t.Errorf("expected user to become verified")
	}
	if updated.AvatarURL != "https://apple.com/avatar.jpg" {
		t.Errorf("expected avatar updated, got %s", updated.AvatarURL)
	}
}

func TestSocialLoginUnsupportedProvider(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	validator := &mockSocialValidator{}
	tokenGen := &mockTokenGen{}

	uc := auth.NewSocialLoginUseCase(repo, validator, tokenGen)

	cmd := auth.SocialLoginCommand{
		Provider: "twitter",
		IDToken:  "some_token",
	}

	_, err := uc.Execute(context.Background(), cmd)
	if !errors.Is(err, user.ErrUnsupportedSocialProvider) {
		t.Fatalf("expected ErrUnsupportedSocialProvider, got %v", err)
	}
}

func TestSocialLoginInvalidToken(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	validator := &mockSocialValidator{err: errors.New("signature expired")}
	tokenGen := &mockTokenGen{}

	uc := auth.NewSocialLoginUseCase(repo, validator, tokenGen)

	cmd := auth.SocialLoginCommand{
		Provider: "google",
		IDToken:  "bad_token",
	}

	_, err := uc.Execute(context.Background(), cmd)
	if !errors.Is(err, user.ErrInvalidSocialToken) {
		t.Fatalf("expected ErrInvalidSocialToken, got %v", err)
	}
}
