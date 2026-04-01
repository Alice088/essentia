package config

import (
	"Alice088/essentia/pkg/currency"
	"time"
)

type Config struct {
	Env           string `env:"ENV,required"`
	StreamManager StreamManager
	LLMManager    LLMManager
}

type StreamManager struct {
	Ticker         time.Duration `env:"SM_TICKER,required"`
	JobBatchCount  int           `env:"SM_JOB_BATCH_COUNT,required"`
	JobPullTimeout time.Duration `env:"SM_JOB_PULL_TIMEOUT,required"`
}

type LLMManager struct {
	SoftBalanceLimit currency.USD `env:"LLM_SOFT_BALANCE_LIMIT" envDefault:"0"`
	MaxBalanceLimit  currency.USD `env:"LLM_MAX_BALANCE_LIMIT" envDefault:"0"`
	Provider         string       `env:"LLM_PROVIDER" envDefault:""`
	ApiKey           string       `env:"LLM_API_KEY" envDefault:""`
	ApiURL           string       `env:"LLM_API_URL" envDefault:""`
}
