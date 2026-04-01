package llm_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/llm"
	"context"
	"sync"
	"testing"
	"time"
)

// mockBalanceProvider is a test double for LLM.
type mockBalanceProvider struct {
	mu          sync.RWMutex
	balance     float64
	err         error
	callCount   int
	lastContext context.Context
}

func (m *mockBalanceProvider) GetBalance(ctx context.Context) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount++
	m.lastContext = ctx
	if m.err != nil {
		return 0, m.err
	}
	return m.balance, nil
}

func (m *mockBalanceProvider) setBalance(balance float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.balance = balance
}

func (m *mockBalanceProvider) setError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func (m *mockBalanceProvider) getCallCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount
}

func (m *mockBalanceProvider) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.callCount = 0
	m.err = nil
	m.lastContext = nil
}

func TestManager_NewWithBalance(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &mockBalanceProvider{balance: 1.0}
	manager := New(cfg, provider)

	if !manager.Enabled() {
		t.Error("Expected manager to be enabled")
	}
	if manager.CurrentBalance() != 0 {
		t.Errorf("Expected initial balance 0, got %f", manager.CurrentBalance())
	}
}

func TestManager_StateTransitions(t *testing.T) {
	tests := []struct {
		name          string
		softLimit     float64
		maxLimit      float64
		balance       float64
		expectedState llm.LimitState
		description   string
	}{
		{
			name:          "normal_above_soft",
			softLimit:     0.50,
			maxLimit:      0.10,
			balance:       1.00,
			expectedState: llm.LimitStateNormal,
			description:   "balance above soft limit → Normal",
		},
		{
			name:          "soft_stop_at_limit",
			softLimit:     0.50,
			maxLimit:      0.10,
			balance:       0.50,
			expectedState: llm.LimitStateSoftStop,
			description:   "balance equals soft limit → SoftStop",
		},
		{
			name:          "soft_stop_below_limit",
			softLimit:     0.50,
			maxLimit:      0.10,
			balance:       0.30,
			expectedState: llm.LimitStateSoftStop,
			description:   "balance below soft limit → SoftStop",
		},
		{
			name:          "max_stop_at_limit",
			softLimit:     0.50,
			maxLimit:      0.10,
			balance:       0.10,
			expectedState: llm.LimitStateMaxStop,
			description:   "balance equals max limit → MaxStop",
		},
		{
			name:          "max_stop_below_limit",
			softLimit:     0.50,
			maxLimit:      0.10,
			balance:       0.05,
			expectedState: llm.LimitStateMaxStop,
			description:   "balance below max limit → MaxStop",
		},
		{
			name:          "no_limits_set",
			softLimit:     0,
			maxLimit:      0,
			balance:       0.01,
			expectedState: llm.LimitStateNormal,
			description:   "No limits configured → Normal",
		},
		{
			name:          "only_soft_limit_normal",
			softLimit:     0.50,
			maxLimit:      0,
			balance:       1.00,
			expectedState: llm.LimitStateNormal,
			description:   "Only soft limit, balance above → Normal",
		},
		{
			name:          "only_soft_limit_stop",
			softLimit:     0.50,
			maxLimit:      0,
			balance:       0.30,
			expectedState: llm.LimitStateSoftStop,
			description:   "Only soft limit, balance below → SoftStop",
		},
		{
			name:          "only_max_limit_normal",
			softLimit:     0,
			maxLimit:      0.10,
			balance:       0.50,
			expectedState: llm.LimitStateNormal,
			description:   "Only max limit, balance above → Normal",
		},
		{
			name:          "only_max_limit_stop",
			softLimit:     0,
			maxLimit:      0.10,
			balance:       0.05,
			expectedState: llm.LimitStateMaxStop,
			description:   "Only max limit, balance below → MaxStop",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LLMManager{
				Enabled:          true,
				SoftBalanceLimit: tt.softLimit,
				MaxBalanceLimit:  tt.maxLimit,
			}
			provider := &mockBalanceProvider{balance: tt.balance}
			manager := New(cfg, provider)

			// UpdateCurrent balance to trigger state calculation
			if err := manager.UpdateBalance(); err != nil {
				t.Fatalf("UpdateBalance failed: %v", err)
			}

			snap := manager.Snapshot()
			if snap.State != tt.expectedState {
				t.Errorf("%s: expected state %s, got %s", tt.description, tt.expectedState, snap.State)
			}
		})
	}
}

func TestManager_Disabled(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          false,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &mockBalanceProvider{balance: 0.05} // Below max limit
	manager := New(cfg, provider)

	snap := manager.Snapshot()
	if snap.State != llm.LimitStateNormal {
		t.Errorf("Disabled manager should return Normal, got %s", snap.State)
	}
	if snap.UsedTokens != 0 {
		t.Errorf("Disabled manager should have 0 used tokens, got %d", snap.UsedTokens)
	}

	// Observe should not affect state
	llmCtx := llm.Context{Prompt: 1000, Completion: 500}
	snap2 := manager.Observe(llmCtx)
	if snap2.State != llm.LimitStateNormal {
		t.Errorf("Observe on disabled manager should return Normal, got %s", snap2.State)
	}
	if snap2.UsedTokens != 0 {
		t.Errorf("Observe on disabled manager should not count tokens, got %d", snap2.UsedTokens)
	}
}

func TestManager_NilManager(t *testing.T) {
	var manager *Manager

	if manager.Enabled() != false {
		t.Error("Nil manager should return false for Enabled()")
	}

	snap := manager.Snapshot()
	if snap.State != llm.LimitStateNormal {
		t.Errorf("Nil manager should return Normal, got %s", snap.State)
	}

	llmCtx := llm.Context{Prompt: 1000, Completion: 500}
	snap2 := manager.Observe(llmCtx)
	if snap2.State != llm.LimitStateNormal {
		t.Errorf("Observe on nil manager should return Normal, got %s", snap2.State)
	}

	if manager.CurrentBalance() != 0 {
		t.Errorf("Nil manager should return 0 balance, got %f", manager.CurrentBalance())
	}

	if err := manager.UpdateBalance(); err != nil {
		t.Errorf("UpdateBalance on nil manager should not error, got %v", err)
	}
}

func TestManager_TokenCounting(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 10.0,
		MaxBalanceLimit:  1.0,
	}
	provider := &mockBalanceProvider{balance: 20.0}
	manager := New(cfg, provider)

	// Initial snapshot
	snap0 := manager.Snapshot()
	if snap0.UsedTokens != 0 {
		t.Errorf("Initial used tokens should be 0, got %d", snap0.UsedTokens)
	}

	// First observation
	llmCtx1 := llm.Context{Prompt: 1000, Completion: 500}
	snap1 := manager.Observe(llmCtx1)
	if snap1.UsedTokens != 1500 {
		t.Errorf("After first observation, used tokens should be 1500, got %d", snap1.UsedTokens)
	}

	// Second observation
	llmCtx2 := llm.Context{Prompt: 200, Completion: 300}
	snap2 := manager.Observe(llmCtx2)
	if snap2.UsedTokens != 2000 {
		t.Errorf("After second observation, used tokens should be 2000, got %d", snap2.UsedTokens)
	}

	// Snapshot should reflect total
	snap3 := manager.Snapshot()
	if snap3.UsedTokens != 2000 {
		t.Errorf("Snapshot should show total used tokens 2000, got %d", snap3.UsedTokens)
	}
}

func TestManager_NoProvider(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	// No provider passed
	manager := New(cfg)

	snap := manager.Snapshot()
	if snap.State != llm.LimitStateNormal {
		t.Errorf("Manager without provider should return Normal, got %s", snap.State)
	}

	// UpdateBalance should do nothing
	if err := manager.UpdateBalance(); err != nil {
		t.Errorf("UpdateBalance without provider should not error, got %v", err)
	}

	if manager.CurrentBalance() != 0 {
		t.Errorf("CurrentBalance without provider should return 0, got %f", manager.CurrentBalance())
	}
}

func TestManager_SetBalanceProvider(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	manager := New(cfg) // No provider initially

	snap1 := manager.Snapshot()
	if snap1.State != llm.LimitStateNormal {
		t.Errorf("Initial state without provider should be Normal, got %s", snap1.State)
	}

	// Set provider with low balance
	provider := &mockBalanceProvider{balance: 0.05}
	manager.SetBalanceProvider(provider)
	if err := manager.UpdateBalance(); err != nil {
		t.Fatalf("UpdateBalance failed: %v", err)
	}

	snap2 := manager.Snapshot()
	if snap2.State != llm.LimitStateMaxStop {
		t.Errorf("After setting provider with low balance, state should be MaxStop, got %s", snap2.State)
	}
}

func TestManager_ProviderError(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &mockBalanceProvider{err: context.Canceled}
	manager := New(cfg, provider)

	// UpdateBalance should return error
	err := manager.UpdateBalance()
	if err == nil {
		t.Error("UpdateBalance should return error when provider fails")
	}
	if err != context.Canceled {
		t.Errorf("Expected error %v, got %v", context.Canceled, err)
	}

	// balance should remain Normal (balance not updated)
	snap := manager.Snapshot()
	if snap.State != llm.LimitStateNormal {
		t.Errorf("balance should remain Normal after provider error, got %s", snap.State)
	}
}

func TestManager_BalanceCaching(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &mockBalanceProvider{balance: 1.0}
	manager := New(cfg, provider)

	// Reduce cache duration for testing
	manager.SetBalanceCacheDuration(100 * time.Millisecond)

	// Initial update
	if err := manager.UpdateBalance(); err != nil {
		t.Fatalf("Initial UpdateBalance failed: %v", err)
	}
	initialCalls := provider.getCallCount()
	if initialCalls != 1 {
		t.Errorf("Expected 1 call after UpdateBalance, got %d", initialCalls)
	}

	// Immediate Snapshot should use cached balance (no new call)
	snap1 := manager.Snapshot()
	callsAfterSnapshot := provider.getCallCount()
	if callsAfterSnapshot != 1 {
		t.Errorf("Snapshot should not call provider (cached), got %d calls", callsAfterSnapshot)
	}
	if snap1.State != llm.LimitStateNormal {
		t.Errorf("balance should be Normal with balance 1.0, got %s", snap1.State)
	}

	// Change provider balance (but cache is fresh)
	provider.setBalance(0.05) // Below max limit
	provider.reset()
	snap2 := manager.Snapshot()
	if provider.getCallCount() != 0 {
		t.Errorf("Cache should prevent call, got %d calls", provider.getCallCount())
	}
	if snap2.State != llm.LimitStateNormal {
		t.Errorf("balance should still be Normal (cached balance 1.0), got %s", snap2.State)
	}

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)
	snap3 := manager.Snapshot()
	if provider.getCallCount() != 1 {
		t.Errorf("After cache expiry, provider should be called once, got %d", provider.getCallCount())
	}
	if snap3.State != llm.LimitStateMaxStop {
		t.Errorf("balance should be MaxStop with balance 0.05, got %s", snap3.State)
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50,
		MaxBalanceLimit:  0.10,
	}
	provider := &mockBalanceProvider{balance: 1.0}
	manager := New(cfg, provider)

	var wg sync.WaitGroup
	goroutines := 10
	iterations := 100

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Mix of operations
				manager.Snapshot()
				if j%3 == 0 {
					llmCtx := llm.Context{Prompt: int64(id*10 + j), Completion: 5}
					manager.Observe(llmCtx)
				}
				if j%7 == 0 {
					manager.CurrentBalance()
				}
				if j%11 == 0 {
					manager.UpdateBalance()
				}
			}
		}(i)
	}
	wg.Wait()

	// Final state should be consistent
	snap := manager.Snapshot()
	if snap.State != llm.LimitStateNormal {
		t.Errorf("Final state should be Normal, got %s", snap.State)
	}
	// No panic indicates thread safety
}
