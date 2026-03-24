package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	UUID      uuid.UUID
	ObjectKey string
}

type ConsumerWorkerPoolConfig struct {
	WorkersCount int
	Timeout      time.Duration
	In           chan Job
	Fn           func(ctx context.Context, in Job, deps *dependencies.AppDeps)
}

func UpConsumerWorkerPool(deps *dependencies.AppDeps, wg *sync.WaitGroup, config ConsumerWorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for range config.WorkersCount {
			go func() {
				defer wg.Done()
				ctx := context.Background()
				for task := range config.In {
					ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
					config.Fn(ctxTimeout, task, deps)
					cancel()
				}
			}()
		}
	}()
}
