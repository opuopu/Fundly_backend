package security

import (
	"context"
	"sync"
)

type MemoryTokenBlacklist struct {
	mu     sync.RWMutex
	tokens map[string]bool
}

func NewMemoryTokenBlacklist() *MemoryTokenBlacklist {
	return &MemoryTokenBlacklist{
		tokens: make(map[string]bool),
	}
}

func (b *MemoryTokenBlacklist) RevokeToken(ctx context.Context, token string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.tokens[token] = true
	return nil
}

func (b *MemoryTokenBlacklist) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.tokens[token], nil
}
