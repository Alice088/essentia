package llm_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/llm"
	"Alice088/essentia/pkg/currency"
	"context"
	"sync"
	"time"
)

type Manager struct {
	enabled bool
	softBal currency.USD
	maxBal  currency.USD

	balanceProvider BalanceProvider

	mu                   sync.Mutex
	totalUsed            llm.Tokens
	currentBalance       currency.USD
	lastBalanceUpdate    time.Time
	balanceCacheDuration time.Duration
}

// New creates a Manager with balance limits (no provider).
func New(cfg config.LLMManager) *Manager {
	return NewWithBalance(cfg, nil)
}

// NewWithBalance creates a Manager with balance limits.
// If provider is nil, balance limits will be ignored.
func NewWithBalance(cfg config.LLMManager, provider BalanceProvider) *Manager {
	return &Manager{
		enabled:              cfg.Enabled,
		softBal:              cfg.SoftBalanceLimit,
		maxBal:               cfg.MaxBalanceLimit,
		balanceProvider:      provider,
		currentBalance:       0,
		balanceCacheDuration: time.Minute,
	}
}

func (m *Manager) Enabled() bool {
	if m == nil {
		return false
	}

	return m.enabled
}

func (m *Manager) Snapshot() llm.Snapshot {
	if m == nil || !m.enabled {
		return llm.Snapshot{State: llm.LimitStateNormal}
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	return llm.Snapshot{
		UsedTokens: m.totalUsed,
		State:      m.stateLocked(),
	}
}

// Observe records consumed tokens from LLM context and returns current state.
func (m *Manager) Observe(ctx llm.Context) llm.Snapshot {
	if m == nil || !m.enabled {
		return llm.Snapshot{State: llm.LimitStateNormal}
	}

	m.mu.Lock()
	m.totalUsed += ctx.Tokens()
	state := m.stateLocked()
	snap := llm.Snapshot{UsedTokens: m.totalUsed, State: state}
	m.mu.Unlock()

	return snap
}

func (m *Manager) SetBalanceProvider(provider BalanceProvider) {
	if m == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balanceProvider = provider
}

func (m *Manager) CurrentBalance() currency.USD {
	if m == nil || m.balanceProvider == nil {
		return 0
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.currentBalance
}

func (m *Manager) UpdateBalance() error {
	if m == nil || m.balanceProvider == nil {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	newBalance, err := m.balanceProvider.GetBalance(context.Background())
	if err != nil {
		return err
	}
	m.currentBalance = newBalance
	m.lastBalanceUpdate = time.Now()
	return nil
}

func (m *Manager) maybeUpdateBalance() {
	if m.balanceProvider == nil {
		return
	}
	if time.Since(m.lastBalanceUpdate) < m.balanceCacheDuration {
		return
	}
	newBalance, err := m.balanceProvider.GetBalance(context.Background())
	if err != nil {
		return
	}
	m.currentBalance = newBalance
	m.lastBalanceUpdate = time.Now()
}

func (m *Manager) stateLocked() llm.LimitState {
	softReached := false
	maxReached := false

	// Check balance limits if configured and provider exists
	if (m.maxBal > 0 || m.softBal > 0) && m.balanceProvider != nil {
		m.maybeUpdateBalance()
		balance := m.currentBalance
		if m.maxBal > 0 && balance <= m.maxBal {
			maxReached = true
		}
		if m.softBal > 0 && balance <= m.softBal {
			softReached = true
		}
	}

	if maxReached {
		return llm.LimitStateMaxStop
	}
	if softReached {
		return llm.LimitStateSoftStop
	}
	return llm.LimitStateNormal
}
