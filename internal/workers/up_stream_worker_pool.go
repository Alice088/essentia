package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

type UpStreamWorkerPoolConfig struct {
	WorkersCount int
	Timeout      time.Duration
	Jobs         chan Job
	GlobalCtx    context.Context
	Worker       func(ctx context.Context, deps dependencies.AppDeps) (Job, error)
	WorkerName   string
}

func UpStreamWorkerPool(deps dependencies.AppDeps, wg *sync.WaitGroup, config UpStreamWorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for i := range config.WorkersCount {
			i++
			deps.Logger.Debug("Starting up worker", "name", config.WorkerName, "worker", i)

			go func() {
				defer deps.Logger.Debug("Worker stop", "name", config.WorkerName, "worker", i)
				defer wg.Done()

				ctx := context.Background()

				for {
					select {
					case <-config.GlobalCtx.Done():
						return
					default:
						ctxTimeout, cancel := context.WithTimeout(ctx, config.Timeout)
						job, err := config.Worker(ctxTimeout, deps)
						cancel()
						if err != nil {
							if !errors.Is(err, pgx.ErrNoRows) {
								deps.Logger.Error("Failed run function", "name", config.WorkerName, "error", err)
							}

							time.Sleep(deps.Config.Workers.StreamJobs.ErrorSleep)
							continue
						}

						select {
						case <-config.GlobalCtx.Done():
							return
						case config.Jobs <- job:
							deps.Logger.Debug("Wrote stream", "name", config.WorkerName, "job_uuid", job.Object.Name.String(), "worker", i)
							continue
						}
					}
				}
			}()
		}
	}()
}
