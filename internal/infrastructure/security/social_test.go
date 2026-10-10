package security_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"

	"github.com/Dodul01/fundly/internal/domain/user"
	"github.com/Dodul01/fundly/internal/infrastructure/security"
)

func TestSocialAuthServiceMockToken(t *testing.T) {
	svc := security.NewSocialAuthService()

	payload, err := svc.ValidateToken(context.Background(), "google", "mock_google_john@example.com")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if payload.Email != "john@example.com" {
		t.Errorf("expected email john@example.com, got %s", payload.Email)
	}
	if payload.Provider != "google" {
		t.Errorf("expected provider google, got %s", payload.Provider)
	}
}

func TestSocialAuthServiceJWTPayload(t *testing.T) {
	svc := security.NewSocialAuthService()

	claims := map[string]interface{}{
		"sub":     "google-sub-999",
		"email":   "oauth.alice@example.com",
		"name":    "Alice OAuth",
		"picture": "https://lh3.googleusercontent.com/alice.png",
	}
	claimsBytes, _ := json.Marshal(claims)
	encodedPayload := base64.RawURLEncoding.EncodeToString(claimsBytes)

	fakeJWT := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9." + encodedPayload + ".fake_signature"

	payload, err := svc.ValidateToken(context.Background(), "google", fakeJWT)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if payload.Email != "oauth.alice@example.com" {
		t.Errorf("expected email oauth.alice@example.com, got %s", payload.Email)
	}
	if payload.FullName != "Alice OAuth" {
		t.Errorf("expected name Alice OAuth, got %s", payload.FullName)
	}
	if payload.AvatarURL != "https://lh3.googleusercontent.com/alice.png" {
		t.Errorf("expected picture url, got %s", payload.AvatarURL)
	}
}

func TestSocialAuthServiceUnsupportedProvider(t *testing.T) {
	svc := security.NewSocialAuthService()

	_, err := svc.ValidateToken(context.Background(), "github", "some_token")
	if !errors.Is(err, user.ErrUnsupportedSocialProvider) {
		t.Fatalf("expected ErrUnsupportedSocialProvider, got %v", err)
	}
}
