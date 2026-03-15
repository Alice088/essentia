package load

import (
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	httpx "Alice088/pdf-summarize/pkg/http"
	"Alice088/pdf-summarize/pkg/size"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

type Handler struct {
	Logger  *slog.Logger
	Queries *queries.Queries
}

func NewHandler(logger *slog.Logger, queries *queries.Queries) Handler {
	return Handler{
		Logger:  logger,
		Queries: queries,
		Timeout
	}
}

func (h *Handler) Load() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

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

		_, err := io.ReadAll(io.LimitReader(r.Body, size.MB5))
		if err != nil {
			if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
				httpx.HttpResponse(w, http.StatusBadRequest, map[string]string{
					"error": "a file too large",
				})
				return
			}

			httpx.HttpResponse(w, http.StatusBadRequest, map[string]string{
				"error": "failed to read file",
			})
			return
		}

		jobID := uuid.New()


		h.Queries.CreateJob()

	}
}
