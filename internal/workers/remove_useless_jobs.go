package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/repo"
	"Alice088/essentia/pkg/env"
	"Alice088/essentia/pkg/s3"
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
	"github.com/jackc/pgx/v5"
)

type UpRemoverUselessJobsWorkerPoolConfig struct {
	WorkersCount int
	Cfg          env.WorkersRemoveUselessJobs
	Timeout      time.Duration
	GlobalCtx    context.Context
	WorkerName   string
}

func UpRemoverUselessJobsWorkerPool(remover RemoverUselessJobs, wg *sync.WaitGroup, config UpRemoverUselessJobsWorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		ticker := time.NewTicker(config.Cfg.Ticker)
		for i := range config.WorkersCount {
			i++
			remover.Logger.Debug("Starting up worker", "name", config.WorkerName, "worker", i)

			go func() {
				defer remover.Logger.Debug("Worker stop", "name", config.WorkerName, "worker", i)
				defer wg.Done()

				ctx := context.Background()

				for {
					select {
					case <-config.GlobalCtx.Done():
						return
					case <-ticker.C:
						ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
						job, err := remover.RemoveUselessJobs(ctxTimeout, deps)
						cancel()
					}
				}
			}()
		}
	}()
}

type RemoverUselessJobs struct {
	Logger *slog.Logger
	Repo   repo.Job
	S3     s3.S3
}

func (p RemoverUselessJobs) RemoveUselessJobs(ctx context.Context, jobs []Job) {
	var objForRemove []s3.Object

	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))
	var mu sync.Mutex
	for _, job := range jobs {
		go func() {
			defer wg.Done()
			ls, err := backoff.Retry[[]s3.Object](ctx, p.listWrapper(ctx, job), backoff.WithMaxTries(3), backoff.WithBackOff(backoff.NewExponentialBackOff()))
			if err != nil {
				return
			}

			for _, object := range ls {
				mu.Lock()
				objForRemove = append(objForRemove, object)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	err := p.S3.DeleteBatch(ctx, objForRemove)
	if err != nil {
		p.Logger.Error(err.Error(), "error", errors.Unwrap(err))
		return
	}
}

func (p RemoverUselessJobs) listWrapper(ctx context.Context, job Job) backoff.Operation[[]s3.Object] {
	i := 1
	return func() ([]s3.Object, error) {
		defer func() { i++ }()
		ls, err := list(ctx, p.S3, job)
		if err != nil {
			p.Logger.Error(err.Error(), "try", i, "error", errors.Unwrap(err))
			return nil, err
		}
		return ls, nil
	}

}

func list(ctx context.Context, storage s3.S3, job Job) ([]s3.Object, error) {
	return storage.List(ctx, s3.ListOptions{Prefix: job.Object.Name.String()})
}
