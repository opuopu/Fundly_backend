package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

func TestResendOTPUseCaseSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "hashed_pass")
	u.SetOTP("111111", 1*time.Minute)

	repo := &mockUserRepo{
		users: map[string]*user.User{
			"usr-1": u,
		},
	}
	otpGen := &mockOTPGen{} // returns "128333"

	uc := auth.NewResendOTPUseCase(repo, otpGen, nil)

	err := uc.Execute(context.Background(), auth.ResendOTPCommand{Email: "user@example.com"})
	if err != nil {
		t.Fatalf("expected resend OTP to succeed, got %v", err)
	}

	savedUser, _ := repo.FindByEmail(context.Background(), "user@example.com")
	if savedUser.OTP != "128333" {
		t.Errorf("expected new OTP 128333, got %s", savedUser.OTP)
	}
}

func TestResendOTPUseCaseUserNotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	uc := auth.NewResendOTPUseCase(repo, &mockOTPGen{}, nil)

	err := uc.Execute(context.Background(), auth.ResendOTPCommand{Email: "unknown@example.com"})
	if err != user.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
