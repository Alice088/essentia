package load

import (
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	httpx "Alice088/pdf-summarize/pkg/http"
	"Alice088/pdf-summarize/pkg/size"
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"
)

type Handler struct {
	Logger     *slog.Logger
	Queries    *queries.Queries
	Timeout    time.Duration
	MinIO      *minio.Client
	BucketName string
}

func NewHandler(logger *slog.Logger, queries *queries.Queries, timeout time.Duration, minio *minio.Client, bucketName string) Handler {
	return Handler{
		Logger:     logger,
		Queries:    queries,
		Timeout:    timeout,
		MinIO:      minio,
		BucketName: bucketName,
	}
}

func (h *Handler) Load() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		if r.ContentLength > size.MB5 {
			httpx.HttpResponse(w, http.StatusRequestEntityTooLarge, map[string]string{
				"error": "file too large",
			})
			return
		}

		header := make([]byte, 5)
		if _, err := io.ReadFull(r.Body, header); err != nil {
			httpx.HttpResponse(w, http.StatusBadRequest, map[string]string{
				"error": "invalid request body",
			})
			return
		}

		if string(header) != "%PDF-" {
			httpx.HttpResponse(w, http.StatusUnsupportedMediaType, map[string]string{
				"error": "invalid pdf file",
			})
			return
		}

		reader := io.MultiReader(
			bytes.NewReader(header),
			r.Body,
		)

		objectSize := r.ContentLength
		if objectSize <= 0 {
			objectSize = -1
			h.Logger.Error("zero-byte size pdf file")
		}

		pdfUUID := uuid.New()

		ctx, cancel := context.WithTimeout(r.Context(), h.Timeout)
		defer cancel()

		_, err := h.MinIO.PutObject(
			ctx,
			h.BucketName,
			pdfUUID.String()+".pdf",
			reader,
			objectSize,
			minio.PutObjectOptions{
				ContentType: "application/pdf",
			},
		)

		if err != nil {
			h.Logger.Error("Failed to put pdf file to minio", "error", err.Error())

			httpx.HttpResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to process file",
			})
			return
		}

		ctx, cancel = context.WithTimeout(r.Context(), h.Timeout)
		defer cancel()

		job, err := h.Queries.CreateJob(ctx, queries.CreateJobParams{
			ID: pgtype.UUID{
				Bytes: pdfUUID,
				Valid: true,
			},
			ObjectKey: pdfUUID.String() + ".pdf",
		})

		if err != nil {
			h.Logger.Error("Failed to create job", "error", err.Error())

			httpx.HttpResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to process file",
			})

			ctx, cancel = context.WithTimeout(r.Context(), h.Timeout)
			defer cancel()
			if err := h.MinIO.RemoveObject(ctx, h.BucketName, pdfUUID.String()+".pdf", minio.RemoveObjectOptions{}); err != nil {
				h.Logger.Error("Failed to remove object during failed create job", "error", err.Error())
			}
			return
		}

		if bytes.Compare(job.ID.Bytes[:], pdfUUID[:]) != 0 {
			h.Logger.Error("Not same job UUID")

			httpx.HttpResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to process file",
			})

			ctx, cancel = context.WithTimeout(r.Context(), h.Timeout)
			defer cancel()
			if err := h.MinIO.RemoveObject(ctx, h.BucketName, pdfUUID.String()+".pdf", minio.RemoveObjectOptions{}); err != nil {
				h.Logger.Error("Failed to remove object during compare job id", "error", err.Error())
			}

			ctx, cancel = context.WithTimeout(r.Context(), h.Timeout)
			defer cancel()
			h.Queries.DeleteJob(ctx, job.ID)
			if err := h.MinIO.RemoveObject(ctx, h.BucketName, pdfUUID.String()+".pdf", minio.RemoveObjectOptions{}); err != nil {
				h.Logger.Error("Failed to remove object during compare job id", "error", err.Error())
			}
			return
		}

		httpx.HttpResponse(w, http.StatusOK, map[string]string{
			"job_id": pdfUUID.String(),
		})
	}
}
