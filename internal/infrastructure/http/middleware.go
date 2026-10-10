package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/Dodul01/fundly/internal/infrastructure/security"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	UserEmailKey contextKey = "user_email"
)

type AuthMiddleware struct {
	jwtService *security.JWTService
	blacklist  *security.MemoryTokenBlacklist
}

func NewAuthMiddleware(jwtService *security.JWTService, blacklist *security.MemoryTokenBlacklist) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
		blacklist:  blacklist,
	}
}

func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Invalid authorization header format, expected Bearer <token>")
			return
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			JSONError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Empty bearer token")
			return
		}

		if m.blacklist != nil {
			if revoked, _ := m.blacklist.IsTokenRevoked(r.Context(), tokenStr); revoked {
				JSONError(w, http.StatusUnauthorized, "TOKEN_REVOKED", "Token has been revoked")
				return
			}
		}

		claims, err := m.jwtService.ValidateToken(tokenStr)
		if err != nil {
			JSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "The provided authentication token is invalid or expired")
			return
		}

		if claims.TokenType != "" && claims.TokenType != "access" {
			JSONError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Expected an access token")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)

		next(w, r.WithContext(ctx))
	}
}

// UserIDFromContext retrieves the user ID from the request context.
func UserIDFromContext(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDKey).(string); ok {
		return val
	}
	return ""
}
