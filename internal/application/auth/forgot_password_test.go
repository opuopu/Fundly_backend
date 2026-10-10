package auth_test

import (
	"context"
	"testing"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

func TestForgotPasswordUseCaseSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hash")
	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	otpGen := &mockOTPGen{}

	uc := auth.NewForgotPasswordUseCase(repo, otpGen)

	err := uc.Execute(context.Background(), auth.ForgotPasswordCommand{
		Email: "user@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	updated, _ := repo.FindByEmail(context.Background(), "user@example.com")
	if updated.OTP != "128333" {
		t.Errorf("expected reset code 128333, got %s", updated.OTP)
	}
}

func TestForgotPasswordUseCaseNonExistentEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	otpGen := &mockOTPGen{}

	uc := auth.NewForgotPasswordUseCase(repo, otpGen)

	// Should succeed without error (silent no-op) to prevent account enumeration
	err := uc.Execute(context.Background(), auth.ForgotPasswordCommand{
		Email: "nonexistent@example.com",
	})
	if err != nil {
		t.Fatalf("expected no error for non-existent email, got %v", err)
	}
}

func TestForgotPasswordUseCaseEmptyEmail(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	otpGen := &mockOTPGen{}

	uc := auth.NewForgotPasswordUseCase(repo, otpGen)

	err := uc.Execute(context.Background(), auth.ForgotPasswordCommand{
		Email: "",
	})
	if err != user.ErrInvalidEmail {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}
