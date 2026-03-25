package workers

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/sqlc"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	errx "Alice088/essentia/pkg/errors"
	"Alice088/essentia/pkg/pdf_reader"
	"Alice088/essentia/pkg/prometheus/metrics"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

// TODO
// - Добавить проверку что если это не первая попытка попытаться взять text из minio (или проверять что есть text_key)
func Parsing(ctx context.Context, job Job, deps *dependencies.AppDeps) {
	logger := deps.Logger.With("uuid=", job.UUID.String(), "stage", "parsing")
	textObjectName := fmt.Sprintf("%s.txt", job.UUID.String())
	start := time.Now()

	var err error
	var pdfSizeBytes int64
	hasPDFSize := false
	defer func() {
		parsingErr := toParsingError(err)
		status := metrics.Success
		if err != nil {
			status = metrics.Failed
			metrics.ParsingErrorsTotal.WithLabelValues(string(parsingErr.Code)).Inc()
		}

		metrics.ParsingDurationSeconds.WithLabelValues(status).Observe(time.Since(start).Seconds())
		if hasPDFSize {
			metrics.ParsingPDFSizeBytes.WithLabelValues(status).Observe(float64(pdfSizeBytes))
		}

		ctxTimeout, cancel := context.WithTimeout(context.Background(), deps.Config.DB.OperationTimeout)
		defer cancel()
		isFailed(ctxTimeout, job, logger, &err, parsingErr, deps)
	}()

	ctxTimeout, cancel := context.WithTimeout(context.Background(), deps.Config.DB.OperationTimeout)
	err = deps.Queries.SetJobStage(ctxTimeout, queries.SetJobStageParams{
		ID:    sqlc.ToUUID(job.UUID),
		Stage: queries.JobStageParsing,
	})
	cancel()
	if err != nil {
		logger.Error("Failed to set job state", "error", err.Error())
		err = errx.NewParsingError(errx.ParsingErrDB, fmt.Errorf("failed to set job state: %w", err))
		return
	}

	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		logger.Error("Failed to create tmp file", "error", err.Error())
		err = errx.NewParsingError(errx.ParsingErrOpen, fmt.Errorf("failed to create tmp file: %w", err))
		return
	}
	defer func() {
		_ = tmpFile.Close()

		tmpErr := os.Remove(tmpFile.Name())
		if tmpErr != nil {
			logger.Error("Failed to cleanup tmp file", "tmp_file", tmpFile.Name(), "error", tmpErr.Error())
		}
	}()

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.MinIO.OperationTimeout)
	err = deps.MinIO.FGetObject(ctxTimeout, "pdf", job.ObjectKey, tmpFile.Name(), minio.GetObjectOptions{})
	cancel()
	if err != nil {
		logger.Error("Failed to get file from minio", "error", err.Error())
		err = errx.NewParsingError(errx.ParsingErrStorageDownload, fmt.Errorf("failed to get file from minio: %w", err))
		return
	}

	tmpFileInfo, tmpStatErr := tmpFile.Stat()
	if tmpStatErr != nil {
		logger.Error("Failed to stat tmp file", "tmp_file", tmpFile.Name(), "error", tmpStatErr.Error())
		err = errx.NewParsingError(errx.ParsingErrOpen, fmt.Errorf("failed to stat tmp file: %w", tmpStatErr))
		return
	}

	pdfSizeBytes = tmpFileInfo.Size()
	hasPDFSize = true

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.Workers.Parsing.ReaderContextTimeout)
	defer cancel()
	res, err := pdf_reader.Read(ctxTimeout, tmpFile.Name())
	if err != nil {
		logger.Error("Failed to read pdf", "error", err.Error())
		err = fmt.Errorf("failed to read pdf: %w", err)
		return
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.MinIO.OperationTimeout)
	_, err = deps.MinIO.PutObject(
		ctxTimeout,
		"pdf",
		textObjectName,
		strings.NewReader(res.Text),
		int64(res.Metadata.Size),
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	cancel()
	if err != nil {
		logger.Error("Failed to put content to minio", "error", err.Error())
		err = errx.NewParsingError(errx.ParsingErrStorageUpload, fmt.Errorf("failed to put content to minio: %w", err))
		return
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
	err = deps.Queries.SetTextKey(ctxTimeout, queries.SetTextKeyParams{
		ID:      sqlc.ToUUID(job.UUID),
		TextKey: sqlc.ToTEXT(textObjectName),
	})
	cancel()
	if err != nil {
		logger.Error("Failed to set text key", "error", err.Error())
		err = errx.NewParsingError(errx.ParsingErrDB, fmt.Errorf("failed to set text key: %w", err))
		return
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
	err = deps.Queries.AdvanceJobStage(ctxTimeout, queries.AdvanceJobStageParams{
		ID:    sqlc.ToUUID(job.UUID),
		Stage: "cleaning",
	})
	cancel()
	if err != nil {
		logger.Error("Failed to advance job", "error", err.Error())
		err = errx.NewParsingError(errx.ParsingErrDB, fmt.Errorf("failed to advance job: %w", err))
		return
	}
}

func isFailed(ctx context.Context, task Job, logger *slog.Logger, err *error, parsingErr *errx.ParsingError, deps *dependencies.AppDeps) {
	if err != nil && *err != nil {
		metrics.ParsingTotal.WithLabelValues(metrics.Failed).Inc()

		dbErr := deps.Queries.FailJob(ctx, queries.FailJobParams{
			ID:    sqlc.ToUUID(task.UUID),
			Error: sqlc.ToTEXT((*err).Error()),
			ErrorType: queries.NullErrorType{
				ErrorType: queries.ErrorType(parsingErr.Code),
				Valid:     true,
			},
		})
		if dbErr != nil {
			logger.Error("Failed to fail job", "error", dbErr.Error())
		}
	} else {
		metrics.ParsingTotal.WithLabelValues(metrics.Success).Inc()
	}
}

func toParsingError(err error) *errx.ParsingError {
	if err == nil {
		return errx.NewParsingError(errx.ParsingErrUnknown, nil)
	}

	if parsingErr, ok := errors.AsType[*errx.ParsingError](err); ok {
		if parsingErr.Code == "" {
			parsingErr.Code = errx.ParsingErrUnknown
		}

		return parsingErr
	}

	return errx.NewParsingError(errx.ParsingErrUnknown, err)
}
