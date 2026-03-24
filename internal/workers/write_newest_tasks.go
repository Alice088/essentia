package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	"context"
	"fmt"
	"sync"
	"time"
)

func WriteNewestTasks(ctx context.Context, deps *dependencies.AppDeps) (Job, error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
	defer cancel()

	job, err := deps.Queries.GetNextJobForStage(ctxTimeout, "uploaded")
	if err != nil {
		return Job{}, fmt.Errorf("failed to get next job: %w", err)
	}

	return Job{
		UUID:      job.ID.Bytes,
		ObjectKey: job.ObjectKey,
	}, nil
}

type WriteNewestTasksWorkerPoolConfig struct {
	WorkersCount int
	Timeout      time.Duration
	Out          chan Job
	GlobalCtx    context.Context
	Fn           func(ctx context.Context, deps *dependencies.AppDeps) (Job, error)
}

func UpWriteNewestTasksWorkerPool(deps *dependencies.AppDeps, wg *sync.WaitGroup, config WriteNewestTasksWorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for range config.WorkersCount {
			go func() {
				defer wg.Done()

				ctx := context.Background()

				for {
					select {
					case <-config.GlobalCtx.Done():
						return
					default:
						ctxTimeout, cancel := context.WithTimeout(ctx, deps.Config.Workers.WriteTasks.ContextTimeout)
						task, err := config.Fn(ctxTimeout, deps)
						cancel()
						if err != nil {
							deps.Logger.Error("Failed make produce", "work", "WriteNewestTasks", "error", err)
							time.Sleep(deps.Config.Workers.WriteTasks.ErrorSleep)
							continue
						}

						select {
						case <-config.GlobalCtx.Done():
							return
						case config.Out <- task:
							continue
						}
					}
				}
			}()
		}
	}()
}
