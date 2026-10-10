package auth

import (
	"context"
)

type TokenBlacklist interface {
	RevokeToken(ctx context.Context, token string) error
	IsTokenRevoked(ctx context.Context, token string) (bool, error)
}

type LogoutCommand struct {
	AccessToken  string
	RefreshToken string
}

type LogoutUseCase struct {
	blacklist TokenBlacklist
}

func NewLogoutUseCase(blacklist TokenBlacklist) *LogoutUseCase {
	return &LogoutUseCase{blacklist: blacklist}
}

func (uc *LogoutUseCase) Execute(ctx context.Context, cmd LogoutCommand) error {
	if uc.blacklist != nil {
		if cmd.AccessToken != "" {
			_ = uc.blacklist.RevokeToken(ctx, cmd.AccessToken)
		}
		if cmd.RefreshToken != "" {
			_ = uc.blacklist.RevokeToken(ctx, cmd.RefreshToken)
		}
	}
	return nil
}
