package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"Alice088/essentia/pkg/s3"
	"context"
	"fmt"
)

func StreamParsingJobs(ctx context.Context, deps dependencies.AppDeps) (job Job, err error) {
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
		Column1: []string{
			string(queries.JobStageUploaded),
			string(queries.JobStageParsing),
		},
		Column2: []string{
			string(queries.ErrorTypeDb),
			string(queries.ErrorTypeStorageUpload),
			string(queries.ErrorTypeStorageDownload),
			string(queries.ErrorTypeUnknown),
		},
	})
	if err != nil {
		err = fmt.Errorf("failed to claim next job: %w", err)
		return
	}

	return Job{
		Object:  s3.MustToPDF(j.ID.String()),
		Attempt: int(j.Attempts),
	}, nil
}
