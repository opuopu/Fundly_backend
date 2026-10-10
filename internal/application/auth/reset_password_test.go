package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

func TestResetPasswordUseCaseSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "old_hash")
	u.SetOTP("489211", 15*time.Minute)

	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	hasher := &mockHasher{}

	uc := auth.NewResetPasswordUseCase(repo, hasher)

	err := uc.Execute(context.Background(), auth.ResetPasswordCommand{
		Email:       "user@example.com",
		ResetCode:   "489211",
		NewPassword: "BrandNewPassword123!",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	updated, _ := repo.FindByEmail(context.Background(), "user@example.com")
	if updated.PasswordHash != "hashed_BrandNewPassword123!" {
		t.Errorf("expected new password hash, got %s", updated.PasswordHash)
	}
	if updated.OTP != "" {
		t.Errorf("expected OTP to be cleared after reset, got %s", updated.OTP)
	}
}

func TestResetPasswordUseCaseInvalidCode(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "old_hash")
	u.SetOTP("489211", 15*time.Minute)

	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	hasher := &mockHasher{}

	uc := auth.NewResetPasswordUseCase(repo, hasher)

	err := uc.Execute(context.Background(), auth.ResetPasswordCommand{
		Email:       "user@example.com",
		ResetCode:   "999999",
		NewPassword: "BrandNewPassword123!",
	})
	if !errors.Is(err, user.ErrInvalidOTP) {
		t.Fatalf("expected ErrInvalidOTP, got %v", err)
	}
}

func TestResetPasswordUseCaseWeakPassword(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "old_hash")
	u.SetOTP("489211", 15*time.Minute)

	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	hasher := &mockHasher{}

	uc := auth.NewResetPasswordUseCase(repo, hasher)

	err := uc.Execute(context.Background(), auth.ResetPasswordCommand{
		Email:       "user@example.com",
		ResetCode:   "489211",
		NewPassword: "short",
	})
	if !errors.Is(err, user.ErrPasswordTooWeak) {
		t.Fatalf("expected ErrPasswordTooWeak, got %v", err)
	}
}

func TestResetPasswordUseCaseUserNotFound(t *testing.T) {
	repo := &mockUserRepo{users: make(map[string]*user.User)}
	hasher := &mockHasher{}

	uc := auth.NewResetPasswordUseCase(repo, hasher)

	err := uc.Execute(context.Background(), auth.ResetPasswordCommand{
		Email:       "ghost@example.com",
		ResetCode:   "489211",
		NewPassword: "BrandNewPassword123!",
	})
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
