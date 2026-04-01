package llm_manager

import (
	"Alice088/essentia/internal/domain/llm"
	"sync"
)

type TokensManager struct {
	total llm.Tokens
	mu    sync.Mutex
}

func NewTokensManager() *TokensManager {
	return &TokensManager{
		total: 0,
		mu:    sync.Mutex{},
	}
}

func (t *TokensManager) Add(tokens llm.Tokens) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.total += tokens
}

func (t *TokensManager) Total() llm.Tokens {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.total
}
