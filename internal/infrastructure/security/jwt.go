package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/Dodul01/fundly/internal/application/auth"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid or expired token")
)

type UserClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	TokenType string `json:"token_type,omitempty"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
}

func NewJWTService(secret string, accessExpiry, refreshExpiry time.Duration) *JWTService {
	if secret == "" {
		secret = "fundly-default-super-secret-key-change-in-prod"
	}
	if accessExpiry <= 0 {
		accessExpiry = 24 * time.Hour
	}
	if refreshExpiry <= 0 {
		refreshExpiry = 7 * 24 * time.Hour
	}

	return &JWTService{
		secretKey:     []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// GenerateTokenPair implements auth.TokenGenerator.
func (s *JWTService) GenerateTokenPair(userID, email string) (*auth.TokenResult, error) {
	now := time.Now().UTC()

	// Access Token
	accessClaims := UserClaims{
		UserID:    userID,
		Email:     email,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpiry)),
			Issuer:    "fundly",
		},
	}
	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err := accessTokenObj.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("signing access token: %w", err)
	}

	// Refresh Token
	refreshClaims := UserClaims{
		UserID:    userID,
		Email:     email,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshExpiry)),
			Issuer:    "fundly",
		},
	}
	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("signing refresh token: %w", err)
	}

	return &auth.TokenResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessExpiry.Seconds()),
	}, nil
}

func (s *JWTService) ValidateToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the userID and email.
func (s *JWTService) ValidateRefreshToken(tokenStr string) (string, string, error) {
	claims, err := s.ValidateToken(tokenStr)
	if err != nil {
		return "", "", err
	}
	if claims.TokenType != "" && claims.TokenType != "refresh" {
		return "", "", ErrInvalidToken
	}
	return claims.UserID, claims.Email, nil
}

