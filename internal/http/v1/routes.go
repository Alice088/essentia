package v1

import (
	"Alice088/pdf-summarize/internal/http/v1/load"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"log/slog"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/minio/minio-go/v7"
)

func Routes(logger *slog.Logger, queries *queries.Queries, timeout time.Duration, minio *minio.Client, bucketName string) chi.Router {
	r := chi.NewRouter()

	loadHandler := load.NewHandler(logger, queries, timeout, minio, bucketName)

	r.With(middleware.AllowContentType("application/pdf")).Post("/load", loadHandler.Load())
	return r
}
