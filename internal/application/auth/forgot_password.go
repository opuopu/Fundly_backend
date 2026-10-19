package auth

import (
	"context"
	"fmt"
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
	mailer Mailer
}

func NewForgotPasswordUseCase(repo user.Repository, otpGen OTPGenerator, mailer Mailer) *ForgotPasswordUseCase {
	return &ForgotPasswordUseCase{
		repo:   repo,
		otpGen: otpGen,
		mailer: mailer,
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

	if err := uc.repo.Update(ctx, u); err != nil {
		return err
	}

	if uc.mailer != nil {
		if err := uc.mailer.SendPasswordReset(ctx, u.Email, resetCode); err != nil {
			return fmt.Errorf("dispatching reset email: %w", err)
		}
	}

	return nil
}
