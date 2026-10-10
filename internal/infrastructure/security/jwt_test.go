package security_test

import (
	"testing"
	"time"

	"github.com/Dodul01/fundly/internal/infrastructure/security"
)

func TestJWTServiceGenerateAndValidate(t *testing.T) {
	svc := security.NewJWTService("secret123", 1*time.Hour, 24*time.Hour)

	pair, err := svc.GenerateTokenPair("usr-1", "user@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validate access token
	accessClaims, err := svc.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Fatalf("expected valid access token, got %v", err)
	}
	if accessClaims.UserID != "usr-1" || accessClaims.Email != "user@example.com" {
		t.Errorf("unexpected claims: %+v", accessClaims)
	}
	if accessClaims.TokenType != "access" {
		t.Errorf("expected token_type access, got %s", accessClaims.TokenType)
	}

	// Validate refresh token
	userID, email, err := svc.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Fatalf("expected valid refresh token, got %v", err)
	}
	if userID != "usr-1" || email != "user@example.com" {
		t.Errorf("unexpected claims: %s, %s", userID, email)
	}

	// Rejection of access token passed to ValidateRefreshToken
	_, _, err = svc.ValidateRefreshToken(pair.AccessToken)
	if err == nil {
		t.Fatalf("expected error when validating access token as refresh token")
	}

	// Rejection of invalid token string
	_, _, err = svc.ValidateRefreshToken("invalid.token.here")
	if err == nil {
		t.Fatalf("expected error for garbage token string")
	}
}
