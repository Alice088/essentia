package app

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/pipeline"
	"Alice088/essentia/internal/llm_manager"
	"Alice088/essentia/internal/llm_provider"
	"Alice088/essentia/internal/stream_manager"
	"Alice088/essentia/pkg/s3"
	"Alice088/essentia/pkg/storage"
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// stubStorage is a minimal implementation of storage.Storage that logs calls.
type stubStorage struct{}

func (s stubStorage) GetProcessableJobs(ctx context.Context, limit int) []storage.Job {
	slog.Warn("stubStorage.GetProcessableJobs called: returning empty list")
	return nil
}

func (s stubStorage) InvalidJob(ctx context.Context, id uuid.UUID) error {
	slog.Warn("stubStorage.InvalidJob called", "job", id.String())
	return nil
}

// stubS3 is a minimal implementation of s3.S3 that logs calls.
type stubS3 struct{}

func (s stubS3) Put(ctx context.Context, file s3.File) error {
	slog.Warn("stubS3.Put called", "name", file.Name, "path", file.Path)
	return nil
}

func (s stubS3) Get(ctx context.Context, file s3.File) error {
	slog.Warn("stubS3.Get called", "name", file.Name, "path", file.Path)
	return nil
}

func (s stubS3) Find(ctx context.Context, name, bucket string) ([]pipeline.Blob, error) {
	slog.Warn("stubS3.Find called", "name", name, "bucket", bucket)
	return nil, nil
}

func (s stubS3) Delete(ctx context.Context, file s3.File) error {
	slog.Warn("stubS3.Delete called", "name", file.Name, "path", file.Path)
	return nil
}

// Run initializes and starts the Essentia application.
func Run(cfg config.Config) {
	logger := slog.Default()
	logger.Info("Starting Essentia", "env", cfg.Env)

	// 1. Initialize LLM Manager with balance limits
	var llmManager *llm_manager.Manager
	logger.Info("LLM Manager enabled",
		"soft_limit", cfg.LLMManager.SoftBalanceLimit,
		"max_limit", cfg.LLMManager.MaxBalanceLimit,
		"provider", cfg.LLMManager.Provider,
	)

	var provider llm_manager.LLM
	switch cfg.LLMManager.Provider {
	case "deepseek":
		if cfg.LLMManager.ApiKey == "" {
			logger.Error("DeepSeek provider selected but LLM_API_KEY not set")
		} else {
			provider = llm_provider.NewDeepSeekProvider(cfg.LLMManager.ApiKey, cfg.LLMManager.ApiURL)
			logger.Info("Created DeepSeek balance provider")
		}
	case "stub", "":
		provider = &llm_manager.StubBalanceProvider{Balance: 100.0}
		logger.Info("Using stub balance provider", "balance", 100.0)
	default:
		logger.Error("Unknown LLM provider", "provider", cfg.LLMManager.Provider)
	}

	llmManager = llm_manager.New(cfg.LLMManager, provider)

	// Perform initial balance update
	if provider != nil {
		if err := llmManager.UpdateBalance(); err != nil {
			logger.Error("Failed to update LLM balance", "err", err)
		} else {
			logger.Info("Initial LLM balance", "balance", llmManager.CurrentBalance())
		}
	}

	// 2. Initialize storage and S3 (using stubs for demonstration)
	// Replace these with real implementations for production use.
	storageImpl := stubStorage{}
	s3Impl := stubS3{}

	// 3. Initialize Stream Manager
	streams := make(map[string]chan pipeline.Job)
	sm := &stream_manager.StreamManager{
		Config:  cfg.StreamManager,
		Streams: streams,
		Storage: storageImpl,
		S3:      s3Impl,
		LLM:     llmManager,
		Logger:  logger,
	}

	// 4. Start Stream Manager in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	sm.Manage(ctx)

	logger.Info("Stream Manager started",
		"ticker", cfg.StreamManager.Ticker,
		"batch_count", cfg.StreamManager.JobBatchCount,
	)

	// 5. Monitor LLM state periodically (example)
	if llmManager != nil {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					snap := llmManager.Snapshot()
					logger.Info("LLM state snapshot",
						"state", snap.State,
						"used_tokens", snap.UsedTokens,
						"balance", llmManager.CurrentBalance(),
					)
				}
			}
		}()
	}

	// 6. Wait for shutdown signal (in real app, catch OS signals)
	logger.Info("Application running. Press Ctrl+C to stop.")
	wg.Wait()
	logger.Info("Application shutdown complete")
}
