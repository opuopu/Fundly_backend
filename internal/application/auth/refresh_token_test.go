package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

type mockRefreshTokenValidator struct {
	userID string
	email  string
	err    error
}

func (m *mockRefreshTokenValidator) ValidateRefreshToken(tokenStr string) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	return m.userID, m.email, nil
}

func TestRefreshTokenUseCaseSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hash")
	u.ClearOTP() // verified

	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	validator := &mockRefreshTokenValidator{userID: "usr-1", email: "user@example.com"}
	tokenGen := &mockTokenGen{}

	uc := auth.NewRefreshTokenUseCase(repo, validator, tokenGen)

	res, err := uc.Execute(context.Background(), auth.RefreshTokenCommand{
		RefreshToken: "valid_refresh_token",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	if res.AccessToken != "mock_access_token" || res.RefreshToken != "mock_refresh_token" {
		t.Errorf("unexpected tokens: %+v", res)
	}
}

func TestRefreshTokenUseCaseInvalidToken(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	validator := &mockRefreshTokenValidator{err: errors.New("token expired")}
	tokenGen := &mockTokenGen{}

	uc := auth.NewRefreshTokenUseCase(repo, validator, tokenGen)

	_, err := uc.Execute(context.Background(), auth.RefreshTokenCommand{
		RefreshToken: "expired_token",
	})
	if !errors.Is(err, user.ErrInvalidRefreshToken) {
		t.Fatalf("expected ErrInvalidRefreshToken, got %v", err)
	}
}

func TestRefreshTokenUseCaseUserNotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	validator := &mockRefreshTokenValidator{userID: "non-existent-user", email: "ghost@example.com"}
	tokenGen := &mockTokenGen{}

	uc := auth.NewRefreshTokenUseCase(repo, validator, tokenGen)

	_, err := uc.Execute(context.Background(), auth.RefreshTokenCommand{
		RefreshToken: "some_token",
	})
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestRefreshTokenUseCaseUserUnverified(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hash")
	u.SetOTP("123456", 5*time.Minute) // Unverified

	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	validator := &mockRefreshTokenValidator{userID: "usr-1", email: "user@example.com"}
	tokenGen := &mockTokenGen{}

	uc := auth.NewRefreshTokenUseCase(repo, validator, tokenGen)

	_, err := uc.Execute(context.Background(), auth.RefreshTokenCommand{
		RefreshToken: "valid_refresh_token",
	})
	if !errors.Is(err, user.ErrAccountNotVerified) {
		t.Fatalf("expected ErrAccountNotVerified, got %v", err)
	}
}
