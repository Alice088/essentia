package llm_provider

import (
	"Alice088/essentia/pkg/currency"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDeepSeekProvider_GetBalance_Success(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/user/balance" {
			t.Errorf("Expected path /user/balance, got %s", r.URL.Path)
		}
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("Expected Authorization header 'Bearer test-api-key', got %s", auth)
		}

		response := map[string]any{
			"is_available": true,
			"balance_infos": []map[string]string{
				{
					"currency":      "USD",
					"total_balance": "125.75",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Fatal(err)
		}
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-api-key", server.URL)
	balance, err := provider.GetBalance(context.Background())
	if err != nil {
		t.Fatalf("GetBalance failed: %v", err)
	}
	expected := 125.75
	if balance != expected {
		t.Errorf("GetBalance() = %v, want %v", balance, expected)
	}
}

func TestDeepSeekProvider_GetBalance_NotAvailable(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"is_available":  false,
			"balance_infos": []map[string]string{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-api-key", server.URL)
	_, err := provider.GetBalance(context.Background())
	if err == nil {
		t.Error("Expected error when balance not available")
	}
	if err.Error() != "balance not available" {
		t.Errorf("Expected error 'balance not available', got %v", err)
	}
}

func TestDeepSeekProvider_GetBalance_EmptyInfos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"is_available":  true,
			"balance_infos": []map[string]string{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-api-key", server.URL)
	_, err := provider.GetBalance(context.Background())
	if err == nil {
		t.Error("Expected error when balance infos empty")
	}
	if err.Error() != "balance info is empty" {
		t.Errorf("Expected error 'balance info is empty', got %v", err)
	}
}

func TestDeepSeekProvider_GetBalance_WrongCurrency(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"is_available": true,
			"balance_infos": []map[string]string{
				{
					"currency":      "EUR",
					"total_balance": "100.00",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-api-key", server.URL)
	_, err := provider.GetBalance(context.Background())
	if err == nil {
		t.Error("Expected error for non-USD currency")
	}
	if err.Error() != "incorrect currency" {
		t.Errorf("Expected error 'incorrect currency', got %v", err)
	}
}

func TestDeepSeekProvider_GetBalance_InvalidNumber(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]any{
			"is_available": true,
			"balance_infos": []map[string]string{
				{
					"currency":      "USD",
					"total_balance": "not-a-number",
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-api-key", server.URL)
	_, err := provider.GetBalance(context.Background())
	if err == nil {
		t.Error("Expected error for invalid number format")
	}
}

func TestDeepSeekProvider_GetBalance_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	provider := NewDeepSeekProvider("test-api-key", server.URL)
	_, err := provider.GetBalance(context.Background())
	if err == nil {
		t.Error("Expected error for HTTP 500")
	}
}

func TestDeepSeekProvider_GetBalance_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Longer than client timeout
	}))
	defer server.Close()

	// Provider with very short timeout
	provider := NewDeepSeekProvider("test-api-key", server.URL)
	provider.client.Timeout = 10 * time.Millisecond

	ctx := context.Background()
	_, err := provider.GetBalance(ctx)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

func TestDeepSeekProvider_DefaultURL(t *testing.T) {
	provider := NewDeepSeekProvider("test-key", "")
	if provider.baseURL != "https://api.deepseek.com" {
		t.Errorf("Expected default URL 'https://api.deepseek.com', got %s", provider.baseURL)
	}
}

func TestDeepSeekProvider_ImplementsInterface(t *testing.T) {
	var _ interface {
		GetBalance(context.Context) (currency.USD, error)
	} = (*DeepSeekProvider)(nil)
	// No compile error means the interface is satisfied
}
