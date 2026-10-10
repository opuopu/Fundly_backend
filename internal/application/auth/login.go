package auth

import (
	"context"
	"fmt"

	"github.com/Dodul01/fundly/internal/domain/user"
)

type PasswordComparator interface {
	Compare(hashedPassword, password string) error
}

type LoginCommand struct {
	Email      string
	Password   string
	RememberMe bool
}

type UserProfile struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type LoginResult struct {
	User   UserProfile `json:"user"`
	Tokens TokenResult `json:"tokens"`
}

type LoginUseCase struct {
	repo       user.Repository
	comparator PasswordComparator
	tokenGen   TokenGenerator
}

func NewLoginUseCase(
	repo user.Repository,
	comparator PasswordComparator,
	tokenGen TokenGenerator,
) *LoginUseCase {
	return &LoginUseCase{
		repo:       repo,
		comparator: comparator,
		tokenGen:   tokenGen,
	}
}

func (uc *LoginUseCase) Execute(ctx context.Context, cmd LoginCommand) (*LoginResult, error) {
	if cmd.Email == "" || cmd.Password == "" {
		return nil, user.ErrInvalidCredentials
	}

	u, err := uc.repo.FindByEmail(ctx, cmd.Email)
	if err != nil {
		return nil, fmt.Errorf("querying user by email: %w", err)
	}
	if u == nil {
		return nil, user.ErrInvalidCredentials
	}

	if err := uc.comparator.Compare(u.PasswordHash, cmd.Password); err != nil {
		return nil, user.ErrInvalidCredentials
	}

	// Verify account has completed OTP registration
	if !u.IsVerified() {
		return nil, user.ErrAccountNotVerified
	}

	tokens, err := uc.tokenGen.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, fmt.Errorf("generating tokens: %w", err)
	}

	return &LoginResult{
		User: UserProfile{
			ID:        u.ID,
			Email:     u.Email,
			FullName:  u.FullName,
			AvatarURL: u.AvatarURL,
		},
		Tokens: *tokens,
	}, nil
}
