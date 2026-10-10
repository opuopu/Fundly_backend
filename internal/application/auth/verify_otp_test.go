package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

type mockTokenGen struct{}

func (m *mockTokenGen) GenerateTokenPair(userID, email string) (*auth.TokenResult, error) {
	return &auth.TokenResult{
		AccessToken:  "mock_access_token",
		RefreshToken: "mock_refresh_token",
		ExpiresIn:    3600,
	}, nil
}

func TestVerifyOTPUseCaseSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	u.SetOTP("128333", 5*time.Minute)

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	tokenGen := &mockTokenGen{}

	uc := auth.NewVerifyOTPUseCase(repo, tokenGen)

	cmd := auth.VerifyOTPCommand{
		Email: "user@example.com",
		OTP:   "128333",
	}

	res, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected verification to succeed, got %v", err)
	}

	if res.User.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", res.User.Email)
	}
	if res.Tokens.AccessToken != "mock_access_token" {
		t.Errorf("expected mock access token, got %s", res.Tokens.AccessToken)
	}

	// Verify OTP was cleared
	savedUser, _ := repo.FindByEmail(context.Background(), "user@example.com")
	if savedUser.OTP != "" {
		t.Errorf("expected OTP to be cleared, got %s", savedUser.OTP)
	}
}

func TestVerifyOTPUseCaseInvalidCode(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	u.SetOTP("128333", 5*time.Minute)

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	uc := auth.NewVerifyOTPUseCase(repo, &mockTokenGen{})

	cmd := auth.VerifyOTPCommand{
		Email: "user@example.com",
		OTP:   "000000",
	}

	_, err := uc.Execute(context.Background(), cmd)
	if err != user.ErrInvalidOTP {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
}
