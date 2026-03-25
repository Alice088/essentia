package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
)

func WriteNewestJobs(ctx context.Context, deps *dependencies.AppDeps) (job Job, err error) {
	ctxTimeout, cancel := context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
	defer cancel()
	tx, err := deps.DB.Begin(ctxTimeout)
	if err != nil {
		err = fmt.Errorf("failed to begin tx: %w", err)
		return
	}
	defer func() {
		if err == nil {
			ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
			defer cancel()
			if err = tx.Commit(ctxTimeout); err != nil {
				err = fmt.Errorf("failed to commit: %w", err)
			} else {
				return
			}
		}

		tx.Rollback(ctx)
	}()

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
	defer cancel()
	j, err := deps.Queries.WithTx(tx).ClaimNextJobForStage(ctxTimeout, queries.ClaimNextJobForStageParams{
		Stage: queries.JobStageUploaded,
		Column2: []queries.ErrorType{
			queries.ErrorTypeDb,
			queries.ErrorTypeStorageUpload,
			queries.ErrorTypeStorageDownload,
			queries.ErrorTypeUnknown,
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to claim next job: %w", err)
		return
	}

	return Job{
		UUID:      j.ID.Bytes,
		ObjectKey: j.ObjectKey,
	}, nil
}

type WriteNewestTasksWorkerPoolConfig struct {
	WorkersCount int
	Timeout      time.Duration
	Jobs         chan Job
	GlobalCtx    context.Context
	Fn           func(ctx context.Context, deps *dependencies.AppDeps) (Job, error)
}

func UpWriteNewestTasksWorkerPool(deps *dependencies.AppDeps, wg *sync.WaitGroup, config WriteNewestTasksWorkerPoolConfig) {
	go func() {
		wg.Add(config.WorkersCount)
		for i := range config.WorkersCount {
			deps.Logger.Debug("Starting up write_newest_jobs worker", "worker", i)

			go func() {
				defer deps.Logger.Debug("Write_newest_jobs worker stop", "worker", i)
				defer wg.Done()

				ctx := context.Background()

				for {
					select {
					case <-config.GlobalCtx.Done():
						return
					default:
						ctxTimeout, cancel := context.WithTimeout(ctx, deps.Config.Workers.WriteTasks.ContextTimeout)
						job, err := config.Fn(ctxTimeout, deps)
						cancel()
						if err != nil {
							if !errors.Is(err, pgx.ErrNoRows) {
								deps.Logger.Error("Failed make produce", "work", "WriteNewestJobs", "error", err)
							}

							time.Sleep(deps.Config.Workers.WriteTasks.ErrorSleep)
							continue
						}

						select {
						case <-config.GlobalCtx.Done():
							return
						case config.Jobs <- job:
							deps.Logger.Debug("Write_newest_jobs write job", "job_uuid", job.UUID.String(), "worker", i)
							continue
						}
					}
				}
			}()
		}
	}()
}
