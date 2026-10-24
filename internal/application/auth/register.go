package auth

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Dodul01/fundly/internal/domain/user"
	"github.com/google/uuid"
)

// Ports (Interfaces) for external dependencies
type PasswordHasher interface {
	Hash(password string) (string, error)
}

type OTPGenerator interface {
	Generate() string
}

type RegisterCommand struct {
	Email   string
	FullName string
	Password string
	ConfirmPassword string
}

type RegisterResult struct {
	UserID string
	Email string
	OTPSent bool
	OTPExpiresInSeconds int
}

type RegisterUseCase struct {
	repo   user.Repository
	hasher PasswordHasher
	otpGen OTPGenerator
	mailer Mailer
}

func NewRegisterUseCase(repo user.Repository, hasher PasswordHasher, otpGen OTPGenerator, mailer Mailer) *RegisterUseCase {
	return &RegisterUseCase{
		repo:   repo,
		hasher: hasher,
		otpGen: otpGen,
		mailer: mailer,
	}
}

func (uc *RegisterUseCase) Execute(ctx context.Context, cmd RegisterCommand) (*RegisterResult, error) {
	// 1. Validation 
	if cmd.Password != cmd.ConfirmPassword {
		return nil, errors.New("passwords do not match")
	}

	if len(cmd.Password) < 8 {
		return nil, user.ErrPasswordTooWeak
	}

		// 2. Check if email already exists
	existingUser, err := uc.repo.FindByEmail(ctx, cmd.Email)
	if existingUser != nil {
		return nil, user.ErrEmailAlreadyExists
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}

	// 3. Hash password
	hashedPassword, err := uc.hasher.Hash(cmd.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 4. Create Domain Entity
	fullName := cmd.FullName
	if fullName == "" {
		parts := strings.Split(cmd.Email, "@")
		if len(parts) > 0 && parts[0] != "" {
			fullName = parts[0]
		} else {
			fullName = "User"
		}
	}

	userID := "usr_" + uuid.New().String()[:8]
	newUser, err := user.New(userID, cmd.Email, fullName, hashedPassword)
	if err != nil {
		return nil, err
	}

	// 5. Set OTP using domain behavior
	otp := uc.otpGen.Generate()
	newUser.SetOTP(otp, 5*time.Minute)

	// 6. Persist
	if err := uc.repo.Save(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	// 7. Dispatch the verification email (best-effort: the OTP is already persisted, so a
	// transient mail failure does not fail registration — the user can request a resend).
	otpSent := true
	if uc.mailer != nil {
		if err := uc.mailer.SendOTP(ctx, newUser.Email, fullName, otp); err != nil {
			log.Printf("register: failed to send verification email to %s: %v", newUser.Email, err)
			otpSent = false
		}
	}

	// 8. Return Result
	return &RegisterResult{
		UserID:              newUser.ID,
		Email:               newUser.Email,
		OTPSent:             otpSent,
		OTPExpiresInSeconds: 300,
	}, nil
}