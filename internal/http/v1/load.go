package v1

import (
	httpx "Alice088/pdf-summarize/pkg/http"
	"Alice088/pdf-summarize/pkg/size"
	"errors"
	"io"
	"log/slog"
	"net/http"
)

func Load(logger *slog.Logger) http.HandlerFunc {
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

		// jobID := uuid.New()

	}
}
