package user

import (
	"context"

	domainUser "github.com/Dodul01/fundly/internal/domain/user"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}

type ChangePasswordInput struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ChangePasswordUseCase struct {
	userRepo domainUser.Repository
	hasher   PasswordHasher
}

func NewChangePasswordUseCase(userRepo domainUser.Repository, hasher PasswordHasher) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo: userRepo,
		hasher:   hasher,
	}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, userID string, input ChangePasswordInput) error {
	if userID == "" {
		return domainUser.ErrInvalidUserData
	}
	if input.CurrentPassword == "" || input.NewPassword == "" {
		return domainUser.ErrInvalidUserData
	}
	if input.NewPassword != input.ConfirmPassword {
		return domainUser.ErrPasswordMismatch
	}
	if len(input.NewPassword) < 8 {
		return domainUser.ErrPasswordTooWeak
	}

	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return domainUser.ErrUserNotFound
	}

	// Verify current password
	if u.PasswordHash == "" {
		return domainUser.ErrInvalidCurrentPassword
	}
	if err := uc.hasher.Compare(u.PasswordHash, input.CurrentPassword); err != nil {
		return domainUser.ErrInvalidCurrentPassword
	}

	// Hash new password
	newHash, err := uc.hasher.Hash(input.NewPassword)
	if err != nil {
		return err
	}

	u.ChangePassword(newHash)

	return uc.userRepo.Update(ctx, u)
}
