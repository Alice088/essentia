package llm_manager

import (
	"Alice088/essentia/pkg/currency"
	"context"
)

type LLM interface {
	// GetBalance returns current balance in USD.
	// Should return an error if balance cannot be retrieved.
	GetBalance(ctx context.Context) (currency.USD, error)
}
