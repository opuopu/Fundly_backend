package auth

import (
	"context"
	"strings"

	"github.com/Dodul01/fundly/internal/domain/user"
)

type ResetPasswordCommand struct {
	Email       string
	NewPassword string
}

type ResetPasswordUseCase struct {
	repo   user.Repository
	hasher PasswordHasher
}

func NewResetPasswordUseCase(repo user.Repository, hasher PasswordHasher) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		repo:   repo,
		hasher: hasher,
	}
}

func (uc *ResetPasswordUseCase) Execute(ctx context.Context, cmd ResetPasswordCommand) error {
	cmd.Email = strings.ToLower(strings.TrimSpace(cmd.Email))

	if cmd.Email == "" {
		return user.ErrInvalidUserData
	}
	if len(cmd.NewPassword) < 8 {
		return user.ErrPasswordTooWeak
	}

	u, err := uc.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return err
	}
	if u == nil {
		return user.ErrUserNotFound
	}

	// Hash new password
	hash, err := uc.hasher.Hash(cmd.NewPassword)
	if err != nil {
		return err
	}

	u.ChangePassword(hash)
	u.ClearOTP()

	return uc.repo.Update(ctx, u)
}
