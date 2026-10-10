package security

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/Dodul01/fundly/internal/domain/user"
)

type SocialAuthService struct{}

func NewSocialAuthService() *SocialAuthService {
	return &SocialAuthService{}
}

// ValidateToken validates a social token and returns identity payload.
func (s *SocialAuthService) ValidateToken(ctx context.Context, provider, idToken string) (*auth.SocialTokenPayload, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "google" && provider != "apple" && provider != "facebook" {
		return nil, user.ErrUnsupportedSocialProvider
	}

	idToken = strings.TrimSpace(idToken)
	if idToken == "" {
		return nil, user.ErrInvalidSocialToken
	}

	// 1. Handle mock/test tokens for dev & integration testing
	if strings.HasPrefix(idToken, "mock_") || strings.HasPrefix(idToken, "test_") {
		email := fmt.Sprintf("%s.user@example.com", provider)
		if strings.Contains(idToken, "@") {
			parts := strings.Split(idToken, "_")
			for _, p := range parts {
				if strings.Contains(p, "@") {
					email = p
					break
				}
			}
		}
		return &auth.SocialTokenPayload{
			Provider:  provider,
			Subject:   fmt.Sprintf("%s_sub_12345", provider),
			Email:     email,
			FullName:  fmt.Sprintf("%s User", strings.ToUpper(provider[:1])+provider[1:]),
			AvatarURL: fmt.Sprintf("https://assets.fundly.com/social/%s.png", provider),
		}, nil
	}

	// 2. If token is in JWT format (3 dot-separated parts), decode claims payload
	parts := strings.Split(idToken, ".")
	if len(parts) == 3 {
		payloadSegment := parts[1]
		if l := len(payloadSegment) % 4; l > 0 {
			payloadSegment += strings.Repeat("=", 4-l)
		}

		decodedBytes, err := base64.URLEncoding.DecodeString(payloadSegment)
		if err == nil {
			var claims struct {
				Sub       string `json:"sub"`
				Email     string `json:"email"`
				Name      string `json:"name"`
				GivenName string `json:"given_name"`
				Picture   string `json:"picture"`
				AvatarURL string `json:"avatar_url"`
			}
			if err := json.Unmarshal(decodedBytes, &claims); err == nil && claims.Email != "" {
				name := claims.Name
				if name == "" {
					name = claims.GivenName
				}
				if name == "" {
					name = strings.Split(claims.Email, "@")[0]
				}

				avatar := claims.Picture
				if avatar == "" {
					avatar = claims.AvatarURL
				}

				return &auth.SocialTokenPayload{
					Provider:  provider,
					Subject:   claims.Sub,
					Email:     claims.Email,
					FullName:  name,
					AvatarURL: avatar,
				}, nil
			}
		}
	}

	return nil, user.ErrInvalidSocialToken
}
