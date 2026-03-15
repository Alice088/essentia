package v1

import (
	"Alice088/pdf-summarize/internal/http/v1/load"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Routes(logger *slog.Logger, queries *queries.Queries) chi.Router {
	r := chi.NewRouter()

	loadHandler := load.NewHandler(logger, queries)

	r.With(middleware.AllowContentType("application/pdf")).Post("/load", loadHandler.Load())
	return r
}
