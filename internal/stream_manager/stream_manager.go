package stream_manager

import (
	"Alice088/essentia/internal/config"
	"Alice088/essentia/internal/domain/llm"
	"Alice088/essentia/internal/domain/pipeline"
	"Alice088/essentia/internal/llm_manager"
	"Alice088/essentia/pkg/s3"
	"Alice088/essentia/pkg/storage"
	"context"
	"log/slog"
	"sync"
	"time"
)

type StreamManager struct {
	Config  config.StreamManager
	Streams map[string]chan pipeline.Job
	Logger  *slog.Logger
	Storage storage.Storage
	S3      s3.S3
	LLM     *llm_manager.Manager

	wg   *sync.WaitGroup
	stop context.CancelFunc
}

func (sm *StreamManager) Manage(ctx context.Context) {
	runCtx, cancel := context.WithCancel(ctx)
	sm.stop = cancel

	go func() {
		if sm.wg != nil {
			defer sm.wg.Done()
		}

		t := time.NewTicker(sm.Config.Ticker)
		defer t.Stop()

		for {
			select {
			case <-runCtx.Done():
				sm.Logger.Info("Stop StreamManager", "cause", runCtx.Err())
				return
			case <-t.C:
				state := sm.llmState()
				if state == llm.LimitStateMaxStop {
					sm.Logger.Error("LLM max token limit reached, emergency stop")
					if sm.stop != nil {
						sm.stop()
					}
					return
				}

				if state == llm.LimitStateSoftStop {
					sm.Logger.Warn("LLM soft token limit reached, waiting pipelines to finish")
					continue
				}

				sm.processTick(runCtx)
			}
		}
	}()
}

func (sm *StreamManager) llmState() llm.LimitState {
	if sm.LLM == nil {
		return llm.LimitStateNormal
	}

	return sm.LLM.Snapshot().State
}

func (sm *StreamManager) processTick(ctx context.Context) {
	jobs := sm.Storage.GetProcessableJobs(ctx, sm.Config.JobBatchCount)
	jobsCh := make(chan pipeline.Job)

	sm.PrepareJobs(jobs, ctx, jobsCh)

	var readyJobs []pipeline.Job
	for job := range jobsCh {
		readyJobs = append(readyJobs, job)
	}

	for _, job := range readyJobs {
		go sm.PullJobs(ctx, job)
	}
}

func (sm *StreamManager) PullJobs(ctx context.Context, job pipeline.Job) {
	timeout, cancel := context.WithTimeout(ctx, sm.Config.JobPullTimeout)
	defer cancel()

	snap := sm.observeLLM(job)

	if snap.State == llm.LimitStateMaxStop {
		sm.Logger.Error("LLM max token limit reached while pushing job", "job", job.JobID.String())
		if sm.stop != nil {
			sm.stop()
		}
		return
	}

	if snap.State == llm.LimitStateSoftStop {
		sm.Logger.Warn("LLM soft token limit reached, skip new dispatch", "job", job.JobID.String())
		return
	}

	stream, ok := sm.Streams[job.Stage]
	if !ok {
		sm.Logger.Error("Stage stream not found", "stage", job.Stage, "job", job.JobID.String())
		return
	}

	select {
	case <-timeout.Done():
		sm.Logger.Warn("Job pull timeout", "job", job.JobID.String())
	case stream <- job:
	}
}

func (sm *StreamManager) observeLLM(job pipeline.Job) llm.Snapshot {
	if sm.LLM == nil {
		return llm.Snapshot{State: llm.LimitStateNormal}
	}

	llmContext := job.LLMContext
	if llmContext.Tokens() == 0 {
		llmContext = llm.Context{
			Prompt:     estimateTokens(len(job.Input)),
			Completion: 0,
		}
	}

	return sm.LLM.Observe(llmContext)
}

func (sm *StreamManager) PrepareJobs(jobs []storage.Job, ctx context.Context, jobsCh chan pipeline.Job) {
	go func() {
		wg := &sync.WaitGroup{}
		wg.Add(len(jobs))

		for _, job := range jobs {
			job := job
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
}

func estimateTokens(blobs int) llm.Tokens {
	if blobs <= 0 {
		return 0
	}

	return int64(blobs) * 256
}
