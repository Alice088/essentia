package config

import "time"

type Config struct {
	Env string `env:"ENV,required"`
}

type SM struct {
	Ticker        time.Duration `env:"SM_TICKER,required"`
	JobBatchCount int           `env:"SM_JOB_BATCH_COUNT,required"`
}
