package auth

import (
	"context"
	"strings"

	"github.com/Dodul01/fundly/internal/domain/user"
)

type RefreshTokenValidator interface {
	ValidateRefreshToken(tokenStr string) (userID, email string, err error)
}

type RefreshTokenCommand struct {
	RefreshToken string
}

type RefreshTokenUseCase struct {
	repo      user.Repository
	validator RefreshTokenValidator
	tokenGen  TokenGenerator
}

func NewRefreshTokenUseCase(
	repo user.Repository,
	validator RefreshTokenValidator,
	tokenGen TokenGenerator,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		repo:      repo,
		validator: validator,
		tokenGen:  tokenGen,
	}
}

func (uc *RefreshTokenUseCase) Execute(ctx context.Context, cmd RefreshTokenCommand) (*TokenResult, error) {
	cmd.RefreshToken = strings.TrimSpace(cmd.RefreshToken)
	if cmd.RefreshToken == "" {
		return nil, user.ErrInvalidRefreshToken
	}

	userID, _, err := uc.validator.ValidateRefreshToken(cmd.RefreshToken)
	if err != nil {
		return nil, user.ErrInvalidRefreshToken
	}

	u, err := uc.repo.GetByID(ctx, userID)
	if err != nil || u == nil {
		return nil, user.ErrUserNotFound
	}

	if !u.IsVerified() {
		return nil, user.ErrAccountNotVerified
	}

	tokens, err := uc.tokenGen.GenerateTokenPair(u.ID, u.Email)
	if err != nil {
		return nil, err
	}

	return tokens, nil
}
