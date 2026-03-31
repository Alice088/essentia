package llm_manager

import (
	"context"
	"testing"
)

func TestStubBalanceProvider_GetBalance(t *testing.T) {
	tests := []struct {
		name     string
		balance  float64
		expected float64
	}{
		{"positive", 10.5, 10.5},
		{"zero", 0.0, 0.0},
		{"negative", -5.0, -5.0},
		{"large", 999999.99, 999999.99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := &StubBalanceProvider{Balance: tt.balance}
			got, err := provider.GetBalance(context.Background())
			if err != nil {
				t.Fatalf("GetBalance returned error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("GetBalance() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStubBalanceProvider_ConcurrentAccess(t *testing.T) {
	provider := &StubBalanceProvider{Balance: 100.0}
	ctx := context.Background()

	const goroutines = 10
	const callsPerGoroutine = 100
	results := make(chan float64, goroutines*callsPerGoroutine)
	errors := make(chan error, goroutines*callsPerGoroutine)

	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < callsPerGoroutine; j++ {
				balance, err := provider.GetBalance(ctx)
				if err != nil {
					errors <- err
					return
				}
				results <- balance
			}
		}()
	}

	// Wait for all goroutines to finish
	received := 0
	for received < goroutines*callsPerGoroutine {
		select {
		case err := <-errors:
			t.Fatalf("Unexpected error: %v", err)
		case balance := <-results:
			if balance != 100.0 {
				t.Errorf("Unexpected balance: %v", balance)
			}
			received++
		}
	}
}
