package workers

import (
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/minio/minio-go/v7"
)

type AntivirusWorker struct {
	UpConfig UpConfig
	Queries  *queries.Queries
	MinIO    *minio.Client
	Logger   *slog.Logger
}

func (aw *AntivirusWorker) Work() {
	jobs := make(chan queries.Job, aw.UpConfig.JobCount)

	go func(jobs chan queries.Job) {
		t := time.NewTicker(aw.UpConfig.JobTick)
		defer t.Stop()

		for {
			select {
			case <-t.C:
				for i := 0; i < aw.UpConfig.JobCount; i++ {
					ctx, cancel := context.WithTimeout(aw.UpConfig.SignalCTX, 2*time.Second)
					job, err := aw.Queries.GetNextUploadedJob(ctx)
					cancel()
					if err != nil {
						if err == sql.ErrNoRows {
							break
						}
						aw.Logger.Error("Failed to get job from databse", "error", err.Error())
						continue
					}
					jobs <- job
				}
			case <-aw.UpConfig.SignalCTX.Done():
				return
			}
		}
	}(jobs)

loop:
	for {
		select {
		case <-jobs:
			//work
		case <-aw.UpConfig.SignalCTX.Done():
			break loop
		}
	}

	<-jobs
}

func (aw *AntivirusWorker) Config() UpConfig {
	return aw.UpConfig
}
