package pdf

import (
	"Alice088/pdf-summarize/internal/dependencies"
	pdfservice "Alice088/pdf-summarize/internal/service"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	httpx "Alice088/pdf-summarize/pkg/http"
	"Alice088/pdf-summarize/pkg/http/pdf"
	"log/slog"
	"net/http"
	"time"

	"github.com/minio/minio-go/v7"
)

type Handler struct {
	Logger     *slog.Logger
	Queries    *queries.Queries
	Timeout    time.Duration
	MinIO      *minio.Client
	PDFService pdfservice.PDFService
}

func NewHandler(appDeps dependencies.AppDeps, serv pdfservice.PDFService) Handler {
	return Handler{
		Logger:     appDeps.Logger,
		Queries:    appDeps.Queries,
		Timeout:    appDeps.Config.HTTP.Timeout,
		MinIO:      appDeps.MinIO,
		PDFService: serv,
	}
}

func (h *Handler) Load() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		result := pdf.BasicValid(r)
		if result.Valid.Error != nil {
			httpx.HttpResponse(w, *result.Valid.Code, map[string]string{
				"error": result.Valid.Error.Error(),
			})
			return
		}

		jobID, err := h.PDFService.CreateJob(
			r.Context(),
			result.Metadata.Reader,
			result.Metadata.Size,
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
