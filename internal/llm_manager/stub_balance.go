package llm_manager

import (
	"Alice088/essentia/pkg/currency"
	"context"
)

// StubBalanceProvider returns a fixed balance for testing.
// Example usage:
//
//	provider := &StubBalanceProvider{balance: 10.0}
//	cfg := config.LLMManager{
//	    Enabled:            true,
//	    SoftBalanceLimit:   0.50,
//	    MaxBalanceLimit:    0.10,
//	    Provider:           "stub",
//	}
//	manager := llm_manager.New(cfg, provider)
//
// Example with DeepSeek provider:
//
//	import "Alice088/essentia/internal/llm_provider"
//	deepseekProvider := llm_provider.NewDeepSeekProvider(
//	    cfg.ApiKey,
//	    cfg.ApiURL,
//	)
//	manager := llm_manager.New(cfg, deepseekProvider)
type StubBalanceProvider struct {
	Balance float64
}

func (s *StubBalanceProvider) GetBalance(ctx context.Context) (currency.USD, error) {
	return s.Balance, nil
}
