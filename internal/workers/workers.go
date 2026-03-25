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
	WorkerName   string
	WorkersCount int
	Timeout      time.Duration
	Jobs         chan Job
	Workers      func(ctx context.Context, in Job, deps *dependencies.AppDeps)
}

func UpConsumerWorkerPool(deps *dependencies.AppDeps, wg *sync.WaitGroup, config ConsumerWorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for i := range config.WorkersCount {
			deps.Logger.Debug("Starting up worker", "name", config.WorkerName, "worker", i)

			go func() {
				defer deps.Logger.Debug("Worker stop", "name", config.WorkerName, "worker", i)
				defer wg.Done()
				ctx := context.Background()
				for task := range config.Jobs {
					deps.Logger.Debug("Process task", "name", config.WorkerName, "task_uuid", task.UUID.String(), "worker", i)
					ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
					config.Workers(ctxTimeout, task, deps)
					cancel()
					deps.Logger.Debug("Finish process task", "name", config.WorkerName, "task_uuid", task.UUID.String(), "worker", i)
				}
			}()
		}
	}()
}
