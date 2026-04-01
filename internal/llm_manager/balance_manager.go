package llm_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/pkg/currency"
	"sync"
	"time"
)

type BalanceManager struct {
	current       currency.USD
	lastUpdate    time.Time
	cacheDuration time.Duration
	policy        BalancePolicy
	mu            sync.Mutex
}

func NewBalanceManager(config config.LLMManager) *BalanceManager {
	return &BalanceManager{
		current: 0,
		policy:  NewBalancePolicy(config),
		mu:      sync.Mutex{},
	}
}

func (b *BalanceManager) UpdateCurrent(new currency.USD) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.current = new
	b.lastUpdate = time.Now()
}

func (b *BalanceManager) Current() currency.USD {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.current
}

func (b *BalanceManager) UpdateCacheDuration(d time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.cacheDuration = d
}

func (b *BalanceManager) IsFresh() bool {
	if b.lastUpdate.IsZero() {
		return false
	}
	return time.Since(b.lastUpdate) < b.cacheDuration
}

func (b *BalanceManager) IsSoftReached() bool {
	return b.policy.IsSoftReached(b)
}

func (b *BalanceManager) IsMaxReached() bool {
	return b.policy.IsMaxReached(b)
}
