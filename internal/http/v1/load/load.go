package load

import (
	"Alice088/pdf-summarize/internal/service"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	httpx "Alice088/pdf-summarize/pkg/http"
	"Alice088/pdf-summarize/pkg/http/pdf"
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
	Service service.PDFService
}

func NewHandler(logger *slog.Logger, queries *queries.Queries, timeout time.Duration, minio *minio.Client) Handler {
	return Handler{
		Logger:  logger,
		Queries: queries,
		Timeout: timeout,
		MinIO:   minio,
	}
}

func (h *Handler) Load() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		safePDF := pdf.BasicValidPDF(r)
		if safePDF.Valid.Error != nil {
			httpx.HttpResponse(w, *safePDF.Valid.Code, map[string]string{
				"error": safePDF.Valid.Error.Error(),
			})
			return
		}

		jobID, err := h.Service.CreateJobFromPDF(
			r.Context(),
			safePDF.Reader,
			safePDF.Size,
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
