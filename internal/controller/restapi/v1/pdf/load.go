package pdf

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/service"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	errs "Alice088/essentia/pkg/errors"
	httpx "Alice088/essentia/pkg/http"
	"Alice088/essentia/pkg/validation"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
)

type Handler struct {
	Logger  *slog.Logger
	Queries *queries.Queries
	Timeout time.Duration
	MinIO   *minio.Client
	Service service.PDF
}

func NewHandler(appDeps *dependencies.AppDeps, serv service.PDF) Handler {
	return Handler{
		Logger:  appDeps.Logger,
		Queries: appDeps.Queries,
		Timeout: appDeps.Config.HTTP.Timeout,
		MinIO:   appDeps.MinIO,
		Service: serv,
	}
}

func (h *Handler) Load() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		pdf := validation.ValidatePDF(validation.PDFInput{
			Size:   r.ContentLength,
			Reader: r.Body,
		})
		if pdf.Error != nil {
			if err, ok := errors.AsType[*errs.PDFError](pdf.Error); !ok {
				httpx.HttpResponse(w, http.StatusInternalServerError, map[string]string{
					"error": "internal server error",
				})
				h.Logger.Error("Failed caught type", "err", err)
			} else {
				httpx.HttpResponse(w, err.StatusCode(), map[string]string{
					"error": err.SafeMessage(),
				})
			}
			return
		}

		jobID, err := h.Service.Enqueue(
			r.Context(),
			pdf.Metadata.Reader,
			pdf.Metadata.Size,
		)
		if err != nil {
			h.Logger.Error("create job failed", "error", err)
			httpx.HttpResponse(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to process file",
			})
			return
		}

		httpx.HttpResponse(w, http.StatusOK, map[string]string{
			"job_id": jobID.String(),
		})
	}
}
