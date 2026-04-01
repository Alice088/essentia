// Package examples contains usage examples for Essentia LLM balance management.
package main

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/llm"
	"Alice088/essentia/internal/llm_manager"
	"Alice088/essentia/internal/llm_provider"
	"log"
	"os"
	"time"
)

func main() {
	log.Println("=== LLM balance Management Example ===")
	config.Load("./.env")

	// Example 1: Using StubBalanceProvider for testing
	log.Println("\n1. Testing with StubBalanceProvider")
	cfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 0.50, // $0.50 USD
		MaxBalanceLimit:  0.10, // $0.10 USD
		Provider:         "stub",
	}
	stubProvider := &llm_manager.StubBalanceProvider{Balance: 1.0} // $1.00 balance
	manager := llm_manager.New(cfg, stubProvider)

	// Check initial state
	snap := manager.Snapshot()
	log.Printf("   Initial state: %s, balance: $%.2f", snap.State, manager.CurrentBalance())

	// Simulate balance dropping below soft limit
	log.Println("\n   Simulating balance drop to $0.30...")
	stubProvider.Balance = 0.30
	manager.UpdateBalance()
	snap = manager.Snapshot()
	log.Printf("   balance after drop: %s, balance: $%.2f", snap.State, manager.CurrentBalance())

	// Simulate balance dropping below max limit
	log.Println("\n   Simulating balance drop to $0.05...")
	stubProvider.Balance = 0.05
	manager.UpdateBalance()
	snap = manager.Snapshot()
	log.Printf("   balance after drop: %s, balance: $%.2f", snap.State, manager.CurrentBalance())

	// Example 2: Using DeepSeekProvider (requires API key)
	log.Println("\n2. Testing with DeepSeekProvider (if API key available)")
	deepseekCfg := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 5.00,
		MaxBalanceLimit:  1.00,
		Provider:         "deepseek",
		ApiKey:           os.Getenv("DEEPSEEK_API_KEY"),
		ApiURL:           os.Getenv("DEEPSEEK_API_URL"),
	}
	if deepseekCfg.ApiKey != "" {
		deepseekProvider := llm_provider.NewDeepSeekProvider(deepseekCfg.ApiKey, deepseekCfg.ApiURL)
		deepseekManager := llm_manager.New(deepseekCfg, deepseekProvider)

		// UpdateCurrent balance from API
		if err := deepseekManager.UpdateBalance(); err != nil {
			log.Printf("   Failed to update balance: %v", err)
		} else {
			log.Printf("   DeepSeek balance: $%.2f", deepseekManager.CurrentBalance())
			snap := deepseekManager.Snapshot()
			log.Printf("   DeepSeek state: %s", snap.State)
		}
	} else {
		log.Println("   Set DEEPSEEK_API_KEY environment variable to test DeepSeek integration")
	}

	// Example 3: Simulating LLM token consumption
	log.Println("\n3. Simulating LLM token consumption with balance limits")
	cfg3 := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 10.00,
		MaxBalanceLimit:  2.00,
		Provider:         "stub",
	}
	stubProvider3 := &llm_manager.StubBalanceProvider{Balance: 15.0}
	manager3 := llm_manager.New(cfg3, stubProvider3)

	// Consume some tokens
	llmCtx := llm.Context{
		Prompt:     1000,
		Completion: 500,
	}
	snap3 := manager3.Observe(llmCtx)
	log.Printf("   After consuming 1,500 tokens: state=%s, total used=%d", snap3.State, snap3.UsedTokens)

	// Drop balance below soft limit
	stubProvider3.Balance = 8.00
	manager3.UpdateBalance()
	snap3 = manager3.Snapshot()
	log.Printf("   balance $8.00 (< soft $10.00): state=%s", snap3.State)

	// Drop balance below max limit
	stubProvider3.Balance = 1.00
	manager3.UpdateBalance()
	snap3 = manager3.Snapshot()
	log.Printf("   balance $1.00 (< max $2.00): state=%s", snap3.State)

	// Example 4: Caching behavior
	log.Println("\n4. Demonstrating balance caching")
	cfg4 := config.LLMManager{
		Enabled:          true,
		SoftBalanceLimit: 5.00,
		MaxBalanceLimit:  1.00,
		Provider:         "stub",
	}
	stubProvider4 := &llm_manager.StubBalanceProvider{Balance: 10.0}
	manager4 := llm_manager.New(cfg4, stubProvider4)
	manager4.UpdateBalance() // initial update

	log.Printf("   Initial balance: $%.2f", manager4.CurrentBalance())
	stubProvider4.Balance = 3.00 // change underlying balance
	log.Println("   Changed stub balance to $3.00 (but cache is still fresh)")

	// stateLocked uses cached balance (still $10.00)
	snap4 := manager4.Snapshot()
	log.Printf("   Cached state: %s (balance still $%.2f)", snap4.State, manager4.CurrentBalance())

	// Wait for cache to expire (default 1 minute)
	log.Println("   Waiting 2 seconds for cache to expire...")
	time.Sleep(2 * time.Second)
	// Change cache duration for demo
	// Note: In real code you would set balanceCacheDuration field
	log.Println("   (In real scenario, cache expires after 1 minute)")

	log.Println("\n=== Example completed ===")
}
