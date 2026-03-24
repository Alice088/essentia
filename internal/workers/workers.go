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
		for i := range config.WorkersCount {
			deps.Logger.Debug("Starting up consumer worker", "worker", i)

			go func() {
				defer deps.Logger.Debug("Consumer worker stop", "worker", i)
				defer wg.Done()
				ctx := context.Background()
				for task := range config.In {
					deps.Logger.Debug("Consumer process task", "task_uuid", task.UUID.String(), "worker", i)
					ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
					config.Fn(ctxTimeout, task, deps)
					cancel()
					deps.Logger.Debug("Consumer finish process task", "task_uuid", task.UUID.String(), "worker", i)
				}
			}()
		}
	}()
}
