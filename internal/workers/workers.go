package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/pkg/s3"
	"context"
	"sync"
	"time"
)

type Job struct {
	Object  s3.Object
	Attempt int
}

type ConsumerWorkerPoolConfig struct {
	WorkerName   string
	WorkersCount int
	Timeout      time.Duration
	Jobs         chan Job
	Workers      func(ctx context.Context, in Job)
}

func UpConsumerWorkerPool(deps dependencies.AppDeps, wg *sync.WaitGroup, config ConsumerWorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for i := range config.WorkersCount {
			i++
			deps.Logger.Debug("Starting up worker", "name", config.WorkerName, "worker", i)

			go func() {
				defer deps.Logger.Debug("Worker stop", "name", config.WorkerName, "worker", i)
				defer wg.Done()
				ctx := context.Background()
				for task := range config.Jobs {
					deps.Logger.Debug("Process task", "name", config.WorkerName, "task_uuid", task.Object.Name.String(), "worker", i)
					ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
					config.Workers(ctxTimeout, task)
					cancel()
					deps.Logger.Debug("Finish process task", "name", config.WorkerName, "task_uuid", task.Object.Name.String(), "worker", i)
				}
			}()
		}
	}()
}
