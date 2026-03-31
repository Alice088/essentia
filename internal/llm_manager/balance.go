package llm_manager

import (
	"Alice088/essentia/pkg/currency"
	"context"
)

// BalanceProvider retrieves current account balance from an LLM provider (e.g., DeepSeek).
// Implement this interface to integrate with your LLM billing API.
type BalanceProvider interface {
	// GetBalance returns current balance in USD.
	// Should return an error if balance cannot be retrieved.
	GetBalance(ctx context.Context) (currency.USD, error)
}
