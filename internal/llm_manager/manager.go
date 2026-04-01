package llm_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/llm"
	"Alice088/essentia/pkg/currency"
	"context"
	"time"
)

type Manager struct {
	llm     LLM
	balance *BalanceManager
	tokens  *TokensManager
}

func New(config config.LLMManager, llm LLM) *Manager {
	return &Manager{
		llm:     llm,
		tokens:  NewTokensManager(),
		balance: NewBalanceManager(config),
	}
}

func (m *Manager) Snapshot() llm.Snapshot {
	return llm.Snapshot{
		UsedTokens: m.tokens.Total(),
		State:      m.state(),
	}
}

func (m *Manager) Observe(ctx llm.Context) llm.Snapshot {
	m.tokens.Add(ctx.Tokens())
	return llm.Snapshot{
		UsedTokens: m.tokens.Total(),
		State:      m.state(),
	}
}

func (m *Manager) CurrentBalance() currency.USD {
	return m.balance.Current()
}

func (m *Manager) UpdateBalance() error {
	balance, err := m.llm.GetBalance(context.Background())
	if err != nil {
		return err
	}
	m.balance.UpdateCurrent(balance)
	return nil
}

func (m *Manager) SetBalanceCacheDuration(cache time.Duration) {
	m.balance.UpdateCacheDuration(cache)
}

func (m *Manager) maybeUpdateBalance() {
	if m.balance.IsFresh() {
		return
	}
	_ = m.UpdateBalance()
}

func (m *Manager) state() llm.LimitState {
	m.maybeUpdateBalance()
	if m.balance.IsMaxReached() {
		return llm.LimitStateMaxStop
	}
	if m.balance.IsSoftReached() {
		return llm.LimitStateSoftStop
	}
	return llm.LimitStateNormal
}
