package workers

import (
	"Alice088/pdf-summarize/internal/app/dependencies"
	"Alice088/pdf-summarize/internal/sqlc"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/ledongthuc/pdf"
	"github.com/minio/minio-go/v7"
)

// TODO
// - Добавить проверку что если это не первая попытка попытаться взять text из minio (или проверять что есть text_key)
func Parsing(ctx context.Context, task Task, deps *dependencies.AppDeps) {
	logger := deps.Logger.With("uuid=", task.UUID.String(), "stage", "parsing")
	textObjectName := fmt.Sprintf("%s-text.txt", task.UUID.String())

	var err error
	defer func() {
		ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		failJob(ctxTimeout, task, logger, &err, deps)
	}()

	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		logger.Error("Failed to create tmp file", "error", err.Error())
		err = fmt.Errorf("failed to create tmp file: %w", err)
		return
	}
	defer func() {
		_ = tmpFile.Close()

		tmpErr := os.Remove(tmpFile.Name())
		if tmpErr != nil {
			logger.Error("Failed to cleanup tmp file", "tmp_file", tmpFile.Name(), "error", tmpErr.Error())
		}
	}()

	ctxTimeout, cancel := context.WithTimeout(ctx, deps.Config.MinIO.OperationTimeout)
	err = deps.MinIO.FGetObject(ctxTimeout, "pdf", task.ObjectKey, tmpFile.Name(), minio.GetObjectOptions{})
	cancel()
	if err != nil {
		logger.Error("Failed to get file from minio", "error", err.Error())
		err = fmt.Errorf("failed to get file from minio: %w", err)
		return
	}

	f, r, err := pdf.Open(tmpFile.Name())
	if err != nil {
		logger.Error("Failed to open file", "error", err.Error())
		err = fmt.Errorf("failed to open file: %w", err)
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		logger.Error("Failed to get plain text", "error", err.Error())
		err = fmt.Errorf("failed to get plain text: %w", err)
		return
	}

	_, err = buf.ReadFrom(b)
	if err != nil {
		logger.Error("Failed to read buffer", "error", err.Error())
		err = fmt.Errorf("failed to read buffer: %w", err)
		return
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.MinIO.OperationTimeout)
	_, err = deps.MinIO.PutObject(
		ctxTimeout,
		"pdf",
		textObjectName,
		&buf,
		int64(buf.Len()),
		minio.PutObjectOptions{
			ContentType: "text/plain",
		},
	)
	cancel()
	if err != nil {
		logger.Error("Failed to put content to minio", "error", err.Error())
		err = fmt.Errorf("failed to put content to minio: %w", err)
		return
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
	err = deps.Queries.SetTextKey(ctxTimeout, queries.SetTextKeyParams{
		ID:      sqlc.ToUUID(task.UUID),
		TextKey: sqlc.ToTEXT(textObjectName),
	})
	cancel()
	if err != nil {
		logger.Error("Failed to set text key", "error", err.Error())
		err = fmt.Errorf("failed to set text key: %w", err)
		return
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, deps.Config.DB.OperationTimeout)
	err = deps.Queries.AdvanceJobStage(ctxTimeout, queries.AdvanceJobStageParams{
		ID:    sqlc.ToUUID(task.UUID),
		Stage: "cleaning",
	})
	cancel()
	if err != nil {
		logger.Error("Failed to advance job", "error", err.Error())
		err = fmt.Errorf("failed to advance job: %w", err)
		return
	}
}

func UpParsingWorkerPool(deps *dependencies.AppDeps, workersCount int, ctx context.Context, tasks <-chan Task) {
	go func() {
		for range workersCount {
			go func() {
				for task := range tasks {
					ctxTimeout, cancel := context.WithTimeout(ctx, deps.Config.Workers.Parsing.ContextTimeout)
					Parsing(ctxTimeout, task, deps)
					cancel()
				}
			}()
		}
	}()

}

func failJob(ctx context.Context, task Task, logger *slog.Logger, err *error, deps *dependencies.AppDeps) {
	if err != nil && *err != nil {
		dbErr := deps.Queries.FailJob(ctx, queries.FailJobParams{
			ID:    sqlc.ToUUID(task.UUID),
			Error: sqlc.ToTEXT((*err).Error()),
		})
		if dbErr != nil {
			logger.Error("Failed to fail job", "error", dbErr.Error())
		}
	}
}
