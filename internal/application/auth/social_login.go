package auth

import (
	"context"
	"strings"

	"github.com/Dodul01/fundly/internal/domain/user"
	"github.com/google/uuid"
)

type SocialTokenPayload struct {
	Provider  string `json:"provider"`
	Subject   string `json:"subject"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	AvatarURL string `json:"avatar_url"`
}

type SocialTokenValidator interface {
	ValidateToken(ctx context.Context, provider, idToken string) (*SocialTokenPayload, error)
}

type SocialLoginCommand struct {
	Provider string
	IDToken  string
}

type SocialLoginUseCase struct {
	repo      user.Repository
	validator SocialTokenValidator
	tokenGen  TokenGenerator
}

func NewSocialLoginUseCase(
	repo user.Repository,
	validator SocialTokenValidator,
	tokenGen TokenGenerator,
) *SocialLoginUseCase {
	return &SocialLoginUseCase{
		repo:      repo,
		validator: validator,
		tokenGen:  tokenGen,
	}
}

func (uc *SocialLoginUseCase) Execute(ctx context.Context, cmd SocialLoginCommand) (*LoginResult, error) {
	cmd.Provider = strings.ToLower(strings.TrimSpace(cmd.Provider))
	cmd.IDToken = strings.TrimSpace(cmd.IDToken)

	if cmd.Provider == "" || cmd.IDToken == "" {
		return nil, user.ErrInvalidUserData
	}

	if cmd.Provider != "google" && cmd.Provider != "apple" && cmd.Provider != "facebook" {
		return nil, user.ErrUnsupportedSocialProvider
	}

	payload, err := uc.validator.ValidateToken(ctx, cmd.Provider, cmd.IDToken)
	if err != nil {
		return nil, user.ErrInvalidSocialToken
	}

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))
	if payload.Email == "" {
		return nil, user.ErrInvalidSocialToken
	}

	existing, err := uc.repo.FindByEmail(ctx, payload.Email)
	if err != nil {
		return nil, err
	}

	var targetUser *user.User
	if existing == nil {
		id := "usr_" + strings.ReplaceAll(uuid.New().String()[:8], "-", "")
		newUser, err := user.NewSocialUser(id, payload.Email, payload.FullName, payload.AvatarURL)
		if err != nil {
			return nil, err
		}
		if err := uc.repo.Save(ctx, newUser); err != nil {
			return nil, err
		}
		targetUser = newUser
	} else {
		needsUpdate := false
		if !existing.IsVerified() {
			existing.ClearOTP()
			needsUpdate = true
		}
		if existing.AvatarURL == "" && payload.AvatarURL != "" {
			existing.SetAvatar(payload.AvatarURL)
			needsUpdate = true
		}
		if existing.FullName == "" && payload.FullName != "" {
			existing.FullName = payload.FullName
			needsUpdate = true
		}
		if needsUpdate {
			if err := uc.repo.Update(ctx, existing); err != nil {
				return nil, err
			}
		}
		targetUser = existing
	}

	tokens, err := uc.tokenGen.GenerateTokenPair(targetUser.ID, targetUser.Email)
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User: UserProfile{
			ID:        targetUser.ID,
			Email:     targetUser.Email,
			FullName:  targetUser.FullName,
			AvatarURL: targetUser.AvatarURL,
		},
		Tokens: *tokens,
	}, nil
}
