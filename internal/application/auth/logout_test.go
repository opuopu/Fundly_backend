package auth_test

import (
	"context"
	"sync"
	"testing"

	"github.com/Dodul01/fundly/internal/application/auth"
)

type mockBlacklist struct {
	mu     sync.Mutex
	tokens map[string]bool
}

func (m *mockBlacklist) RevokeToken(_ context.Context, token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokens[token] = true
	return nil
}

func (m *mockBlacklist) IsTokenRevoked(_ context.Context, token string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.tokens[token], nil
}

func TestLogoutUseCase(t *testing.T) {
	bl := &mockBlacklist{tokens: make(map[string]bool)}
	uc := auth.NewLogoutUseCase(bl)

	err := uc.Execute(context.Background(), auth.LogoutCommand{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	revoked, _ := bl.IsTokenRevoked(context.Background(), "access-token-123")
	if !revoked {
		t.Errorf("expected access token to be revoked")
	}

	revokedRefresh, _ := bl.IsTokenRevoked(context.Background(), "refresh-token-456")
	if !revokedRefresh {
		t.Errorf("expected refresh token to be revoked")
	}
}
