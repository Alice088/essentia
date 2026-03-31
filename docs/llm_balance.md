# LLM Balance Management

Essentia now supports LLM cost control based on account balance (USD/USDT) instead of token counts. This prevents unexpected costs when using paid LLM APIs like DeepSeek.

## Configuration

Add these environment variables to your `.env` file:

```bash
# Enable LLM Manager
LLM_MANAGER_ENABLED=true

# Balance limits (USD)
LLM_SOFT_BALANCE_LIMIT=0.50    # Soft stop at $0.50
LLM_MAX_BALANCE_LIMIT=0.10     # Emergency stop at $0.10

# Provider configuration (DeepSeek example)
LLM_PROVIDER=deepseek          # "deepseek" or "stub"
LLM_API_KEY=your_deepseek_api_key
LLM_API_URL=https://api.deepseek.com  # Optional, defaults to DeepSeek API
```

## How It Works

### States
- **Normal**: Balance above soft limit → jobs processed normally
- **Soft Stop**: Balance ≤ soft limit → new jobs are **not** dispatched, existing pipelines finish
- **Max Stop**: Balance ≤ max limit → emergency shutdown of StreamManager

### Integration Points
1. **StreamManager** checks LLM state every tick:
   - `LimitStateSoftStop`: Skips `processTick()`, no new jobs pulled
   - `LimitStateMaxStop`: Calls `sm.stop()`, emergency shutdown
2. **PullJobs** verifies state before dispatching:
   - Returns early if `LimitStateSoftStop` or `LimitStateMaxStop`

### Balance Providers
Implement the `BalanceProvider` interface to integrate with any LLM billing API:

```go
type BalanceProvider interface {
    GetBalance(ctx context.Context) (float64, error)
}
```

Built-in providers:
- **StubBalanceProvider**: Fixed balance for testing
- **DeepSeekProvider**: Fetches balance from DeepSeek API (assumes `/v1/balance` endpoint)

## Usage Examples

### Basic Setup
```go
import (
    "Alice088/essentia/internal/config"
    "Alice088/essentia/internal/llm_manager"
)

cfg := config.LLMManager{
    Enabled:          true,
    SoftBalanceLimit: 0.50,
    MaxBalanceLimit:  0.10,
    Provider:         "stub",
}

provider := &llm_manager.StubBalanceProvider{Balance: 1.0}
manager := llm_manager.NewWithBalance(cfg, provider)

// Check state
snap := manager.Snapshot()
fmt.Printf("State: %s, Balance: $%.2f\n", snap.State, manager.CurrentBalance())
```

### DeepSeek Integration
```go
import "Alice088/essentia/internal/llm_provider"

provider := llm_provider.NewDeepSeekProvider(apiKey, apiURL)
manager := llm_manager.NewWithBalance(cfg, provider)

// Update balance (auto-cached for 1 minute)
err := manager.UpdateBalance()
```

### With StreamManager
```go
sm := &stream_manager.StreamManager{
    Config:  streamCfg,
    Storage: storageImpl,
    S3:      s3Impl,
    LLM:     manager,  // Your LLM manager with balance limits
    Logger:  logger,
}

// Start managing streams
sm.Manage(ctx)
```

## Monitoring

The manager maintains:
- **Used tokens**: Total tokens consumed (for monitoring)
- **Current balance**: Cached balance (updated every minute)
- **State**: Current limit state

Monitor via `Snapshot()` or periodic logging:

```go
snap := manager.Snapshot()
log.Printf("State: %s, Used: %d tokens, Balance: $%.2f",
    snap.State, snap.UsedTokens, manager.CurrentBalance())
```

## Cache Behavior

Balance queries are cached for **1 minute** (configurable via `balanceCacheDuration` field). This prevents excessive API calls while maintaining timely limit enforcement.

## Testing

Use `StubBalanceProvider` for deterministic tests:

```go
provider := &llm_manager.StubBalanceProvider{Balance: 10.0}
manager := llm_manager.NewWithBalance(cfg, provider)

// Test soft limit
provider.Balance = 0.30  // Below soft limit of 0.50
manager.UpdateBalance()
assert.Equal(t, llm.LimitStateSoftStop, manager.Snapshot().State)
```

## Environment Variables Reference

| Variable | Default | Description |
|----------|---------|-------------|
| `LLM_MANAGER_ENABLED` | `false` | Enable LLM balance management |
| `LLM_SOFT_BALANCE_LIMIT` | `0` | Soft stop balance (USD), 0 = disabled |
| `LLM_MAX_BALANCE_LIMIT` | `0` | Emergency stop balance (USD), 0 = disabled |
| `LLM_PROVIDER` | `""` | `"deepseek"`, `"stub"`, or custom |
| `LLM_API_KEY` | `""` | API key for provider (DeepSeek) |
| `LLM_API_URL` | `""` | Optional provider API URL |

## Migration from Token Limits

Previous token-based limits (`LLM_SOFT_TOKEN_LIMIT`, `LLM_MAX_TOKEN_LIMIT`) have been removed. The system now uses only balance limits.

Token counting remains active for monitoring (`UsedTokens` in `Snapshot`), but does not affect limits.