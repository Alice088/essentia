package steam_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/pipeline"
	"Alice088/essentia/pkg/s3"
	"Alice088/essentia/pkg/storage"
	"context"
	"log/slog"
	"sync"
	"time"
)

type StreamManager struct {
	Config  config.SM
	Streams map[string]chan pipeline.Job
	Logger  *slog.Logger
	Storage storage.Storage
	S3      s3.S3
	wg      *sync.WaitGroup
}

func (sm StreamManager) Managing(ctx context.Context) {
	go func() {
		defer sm.wg.Done()
		var jobs []storage.Job
		sem := make(chan struct{}, 1)

		t := time.NewTicker(sm.Config.Ticker)

		select {
		case <-ctx.Done():
			sm.Logger.Info("Stop SM", "cause", ctx.Err())
		case <-t.C:
			sem <- struct{}{}
			jobs = sm.Storage.GetProcessableJobs(ctx, sm.Config.JobBatchCount)

			var readyJobs []pipeline.Job
			jobsCh := make(chan pipeline.Job)

			jobsCh = sm.PrepareJobs(jobs, ctx, jobsCh)

			for job := range jobsCh {
				readyJobs = append(readyJobs, job)
			}

			for _, job := range readyJobs {
				go sm.PullJobs(ctx, job)
			}
		}
	}()
}

func (sm StreamManager) PullJobs(ctx context.Context, job pipeline.Job) {
	timeout, cancel := context.WithTimeout(ctx, sm.Config.JobPullTimeout)
	defer cancel()

	select {
	case <-timeout.Done():
	case sm.Streams[job.Stage] <- job:
	}
}

func (sm StreamManager) PrepareJobs(jobs []storage.Job, ctx context.Context, jobsCh chan pipeline.Job) chan pipeline.Job {
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(len(jobs))

		for _, job := range jobs {
			go func() {
				defer wg.Done()
				find, err := sm.S3.Find(ctx, job.ID.String(), string(job.Stage))
				if err != nil {
					sm.Logger.Error("Failed to find blob", "job", job.ID.String(), "bucket", string(job.Stage))
					err = sm.Storage.InvalidJob(ctx, job.ID)
					if err != nil {
						sm.Logger.Error("Failed to invalid job", "job", job.ID.String())
					}
					return
				}

				jobsCh <- pipeline.Job{
					JobID: job.ID,
					Input: find,
					Stage: string(job.Stage),
				}
			}()
		}

		wg.Wait()
		close(jobsCh)
	}()
	return jobsCh
}
