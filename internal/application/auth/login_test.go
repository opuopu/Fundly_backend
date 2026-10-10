package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

type mockComparator struct{}

func (m *mockComparator) Compare(hash, password string) error {
	if hash == "hashed_"+password {
		return nil
	}
	return errors.New("mismatched password")
}

func TestLoginUseCaseSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Thom Haye", "hashed_Password123!")
	u.ClearOTP() // verified

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	comparator := &mockComparator{}
	tokenGen := &mockTokenGen{}

	uc := auth.NewLoginUseCase(repo, comparator, tokenGen)

	cmd := auth.LoginCommand{
		Email:      "user@example.com",
		Password:   "Password123!",
		RememberMe: true,
	}

	res, err := uc.Execute(context.Background(), cmd)
	if err != nil {
		t.Fatalf("expected login to succeed, got %v", err)
	}

	if res.User.Email != "user@example.com" {
		t.Errorf("expected email user@example.com, got %s", res.User.Email)
	}
	if res.Tokens.AccessToken != "mock_access_token" {
		t.Errorf("expected mock access token, got %s", res.Tokens.AccessToken)
	}
}

func TestLoginUseCaseWrongPassword(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Thom Haye", "hashed_Password123!")
	u.ClearOTP()

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	uc := auth.NewLoginUseCase(repo, &mockComparator{}, &mockTokenGen{})

	cmd := auth.LoginCommand{
		Email:    "user@example.com",
		Password: "IncorrectPassword",
	}

	_, err := uc.Execute(context.Background(), cmd)
	if err != user.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUseCaseUnverifiedAccount(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Thom Haye", "hashed_Password123!")
	u.SetOTP("128333", 5*time.Minute) // unverified

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	uc := auth.NewLoginUseCase(repo, &mockComparator{}, &mockTokenGen{})

	cmd := auth.LoginCommand{
		Email:    "user@example.com",
		Password: "Password123!",
	}

	_, err := uc.Execute(context.Background(), cmd)
	if err != user.ErrAccountNotVerified {
		t.Fatalf("expected ErrAccountNotVerified, got %v", err)
	}
}
