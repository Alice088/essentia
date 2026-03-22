package workers

import (
	"Alice088/pdf-summarize/internal/dependencies"
	"Alice088/pdf-summarize/internal/sqlc"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/ledongthuc/pdf"
	"github.com/minio/minio-go/v7"
)

// TODO
// - Добавить проверку что если это не первая попытка попытаться взять text из minio (или проверять что есть text_key)
func (w *Worker) Parsing(ctx context.Context, task Task) {
	logger := w.Logger.With("uuid=", task.UUID.String(), "stage", "parsing")
	textObjectName := fmt.Sprintf("%s-text.txt", task.UUID.String())

	var err error
	defer w.failJob(ctx, logger, task, &err)

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

	ctxTimeout, cancel := context.WithTimeout(ctx, w.Config.ContextTimeout)
	err = w.MinIO.FGetObject(ctxTimeout, "pdf", task.ObjectKey, tmpFile.Name(), minio.GetObjectOptions{})
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

	ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
	_, err = w.MinIO.PutObject(
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

	ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
	err = w.Queries.SetTextKey(ctxTimeout, queries.SetTextKeyParams{
		ID:      sqlc.ToUUID(task.UUID),
		TextKey: sqlc.ToTEXT(textObjectName),
	})
	cancel()
	if err != nil {
		logger.Error("Failed to set text key", "error", err.Error())
		err = fmt.Errorf("failed to set text key: %w", err)
		return
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
	err = w.Queries.AdvanceJobStage(ctxTimeout, queries.AdvanceJobStageParams{
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

func UpParsingWorkerPool(deps dependencies.AppDeps, workersCount int, ctx context.Context, tasks <-chan Task) {
	go func() {
		for range workersCount {
			go func() {
				w := NewWorker(deps)
				for task := range tasks {
					w.Parsing(ctx, task)
				}
			}()
		}
	}()

}

func (w *Worker) failJob(ctx context.Context, logger *slog.Logger, task Task, err *error) {
	if err != nil && *err != nil {
		ctxTimeout, cancel := context.WithTimeout(ctx, w.Config.ContextTimeout)
		dbErr := w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
			ID:    sqlc.ToUUID(task.UUID),
			Error: sqlc.ToTEXT((*err).Error()),
		})
		cancel()
		if dbErr != nil {
			logger.Error("Failed to fail job", "error", dbErr.Error())
		}
	}
}
