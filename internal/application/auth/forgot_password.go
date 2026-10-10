package auth

import (
	"context"
	"strings"
	"time"

	"github.com/Dodul01/fundly/internal/domain/user"
)

type ForgotPasswordCommand struct {
	Email string
}

type ForgotPasswordUseCase struct {
	repo   user.Repository
	otpGen OTPGenerator
}

func NewForgotPasswordUseCase(repo user.Repository, otpGen OTPGenerator) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		repo:   repo,
		otpGen: otpGen,
	}
}

func (uc *ForgotPasswordUseCase) Execute(ctx context.Context, cmd ForgotPasswordCommand) error {
	email := strings.ToLower(strings.TrimSpace(cmd.Email))
	if email == "" {
		return user.ErrInvalidEmail
	}

	u, err := uc.repo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}
	if u == nil {
		// Return nil silently to prevent email enumeration attack
		return nil
	}

	resetCode := uc.otpGen.Generate()
	u.SetOTP(resetCode, 15*time.Minute)

	return uc.repo.Update(ctx, u)
}
