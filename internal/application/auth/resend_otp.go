package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Dodul01/fundly/internal/domain/user"
)

type ResendOTPCommand struct {
	Email string
}

type ResendOTPUseCase struct {
	repo   user.Repository
	otpGen OTPGenerator
}

func NewResendOTPUseCase(repo user.Repository, otpGen OTPGenerator) *ResendOTPUseCase {
	return &ResendOTPUseCase{
		repo:   repo,
		otpGen: otpGen,
	}
}

func (uc *ResendOTPUseCase) Execute(ctx context.Context, cmd ResendOTPCommand) error {
	if cmd.Email == "" {
		return user.ErrInvalidUserData
	}

	u, err := uc.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return fmt.Errorf("finding user by email: %w", err)
	}
	if u == nil {
		return user.ErrUserNotFound
	}

	newOTP := uc.otpGen.Generate()
	u.SetOTP(newOTP, 5*time.Minute)

	if err := uc.repo.Save(ctx, u); err != nil {
		return fmt.Errorf("persisting refreshed otp: %w", err)
	}

	return nil
}
