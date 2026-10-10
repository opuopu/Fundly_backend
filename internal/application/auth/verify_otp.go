package auth

import (
	"context"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/user"
)

type TokenGenerator interface {
	GenerateTokenPair(userID, email string) (*TokenResult, error)
}

type TokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type VerifyOTPCommand struct {
	Email string
	OTP   string
}

type UserSummary struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type VerifyOTPResult struct {
	User   UserSummary `json:"user"`
	Tokens TokenResult `json:"tokens"`
}

type VerifyOTPUseCase struct {
	repo     user.Repository
	tokenGen TokenGenerator
}

func NewVerifyOTPUseCase(repo user.Repository, tokenGen TokenGenerator) *VerifyOTPUseCase {
	return &VerifyOTPUseCase{
		repo:     repo,
		tokenGen: tokenGen,
	}
}

func (uc *VerifyOTPUseCase) Execute(ctx context.Context, cmd VerifyOTPCommand) (*VerifyOTPResult, error) {
	if cmd.Email == "" || cmd.OTP == "" {
		return nil, user.ErrInvalidUserData
	}

	u, err := uc.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("finding user by email: %w", err)
	}
	if u == nil {
		return nil, user.ErrUserNotFound
	}

	if err := u.VerifyOTP(cmd.OTP); err != nil {
		return nil, err
	}

	u.ClearOTP()
	if err := uc.repo.Save(ctx, u); err != nil {
		return nil, fmt.Errorf("clearing otp on user: %w", err)
	}

	tokens, err := uc.tokenGen.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, fmt.Errorf("generating authentication tokens: %w", err)
	}

	return &VerifyOTPResult{
		User: UserSummary{
			ID:    u.ID,
			Email: u.Email,
		},
		Tokens: *tokens,
	}, nil
}
