package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

func TestResetPasswordUseCaseSuccess(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "old_hash")

	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	hasher := &mockHasher{}

	uc := auth.NewResetPasswordUseCase(repo, hasher)

	err := uc.Execute(context.Background(), auth.ResetPasswordCommand{
		Email:       "user@example.com",
		NewPassword: "BrandNewPassword123!",
	})
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	updated, _ := repo.FindByEmail(context.Background(), "user@example.com")
	if updated.PasswordHash != "hashed_BrandNewPassword123!" {
		t.Errorf("expected new password hash, got %s", updated.PasswordHash)
	}
}

func TestResetPasswordUseCaseWeakPassword(t *testing.T) {
	u, _ := user.New("usr-1", "user@example.com", "Test User", "old_hash")

	repo := &mockUserRepo{
		users: map[string]*user.User{"usr-1": u},
	}
	hasher := &mockHasher{}

	uc := auth.NewResetPasswordUseCase(repo, hasher)

	err := uc.Execute(context.Background(), auth.ResetPasswordCommand{
		Email:       "user@example.com",
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
		NewPassword: "BrandNewPassword123!",
	})
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
