package workers

import (
	"Alice088/pdf-summarize/internal/sqlc"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/minio/minio-go/v7"
)

// TODO
// - Добавить проверку что если это не первая попытка попытаться взять text из minio (или проверять что есть text_key)
func (w *Worker) Parsing(ctx context.Context, tasks <-chan Task) {
	l := w.Logger.With("stage=parsing")

	for task := range tasks {
		logger := l.With("uuid=", task.UUID.String())

		tmpFile, err := os.CreateTemp("", "*.pdf")
		if err != nil {
			logger.Error("Failed to create tmp file", "error", err.Error())

			ctxTimeout, cancel := context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to create tmp file: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			return
		}

		ctxTimeout, cancel := context.WithTimeout(ctx, w.Config.ContextTimeout)
		err = w.MinIO.FGetObject(ctxTimeout, "pdf", task.ObjectKey, tmpFile.Name(), minio.GetObjectOptions{})
		cancel()

		if err != nil {
			logger.Error("Failed to get file from minio", "error", err.Error())

			ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to get file: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			return
		}

		f, r, err := pdf.Open(tmpFile.Name())
		if err != nil {
			logger.Error("Failed to open file", "error", err.Error())

			ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to open file: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			err = f.Close()
			if err != nil {
				logger.Error("Failed to close file", "error", err.Error())
			}
			return
		}

		var buf bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			logger.Error("Failed to get plain text", "error", err.Error())

			ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to get plain text: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			err = f.Close()
			if err != nil {
				logger.Error("Failed to close file", "error", err.Error())
			}
			return
		}

		_, err = buf.ReadFrom(b)
		if err != nil {
			logger.Error("Failed to read buffer", "error", err.Error())

			ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to read buffer: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			err = f.Close()
			if err != nil {
				logger.Error("Failed to close file", "error", err.Error())
			}
			return
		}

		content := buf.String()
		textObjectName := fmt.Sprintf("%s-text.txt", task.UUID.String())
		ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
		_, err = w.MinIO.PutObject(
			ctx,
			"pdf",
			textObjectName,
			strings.NewReader(content),
			int64(len(content)),
			minio.PutObjectOptions{
				ContentType: "text/plain",
			},
		)
		cancel()
		if err != nil {
			logger.Error("Failed to put content to minio", "error", err.Error())

			ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to put content to minio: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			err = f.Close()
			if err != nil {
				logger.Error("Failed to close file", "error", err.Error())
			}
			return
		}

		ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
		err = w.Queries.SetTextKey(ctx, queries.SetTextKeyParams{
			ID:      sqlc.ToUUID(task.UUID),
			TextKey: sqlc.ToTEXT(textObjectName),
		})
		cancel()
		if err != nil {
			logger.Error("Failed to set text key", "error", err.Error())

			ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to set text key: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			err = f.Close()
			if err != nil {
				logger.Error("Failed to close file", "error", err.Error())
			}
			return
		}

		ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
		err = w.Queries.AdvanceJobStage(ctx, queries.AdvanceJobStageParams{
			ID:    sqlc.ToUUID(task.UUID),
			Stage: "cleaning",
		})
		cancel()
		if err != nil {
			logger.Error("Failed to advance job", "error", err.Error())

			ctxTimeout, cancel = context.WithTimeout(ctx, w.Config.ContextTimeout)
			err = w.Queries.FailJob(ctxTimeout, queries.FailJobParams{
				ID:    sqlc.ToUUID(task.UUID),
				Error: sqlc.ToTEXT("Failed to advance job: " + err.Error()),
			})
			cancel()

			if err != nil {
				logger.Error("Failed to fail job :0", "error", err.Error())
			}
			err = f.Close()
			if err != nil {
				logger.Error("Failed to close file", "error", err.Error())
			}
			return
		}

		err = f.Close()
		if err != nil {
			logger.Error("Failed to close file", "error", err.Error())
		}

		err = os.Remove(tmpFile.Name())
		if err != nil {
			logger.Error("Failed to cleanup tmp file", "tmp_file=", tmpFile.Name(), "error", err.Error())
		}
	}
}
