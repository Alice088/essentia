package v1

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Routes(logger *slog.Logger) chi.Router {
	r := chi.NewRouter()

	r.With(middleware.AllowContentType("application/pdf")).Post("/load", Load(logger))
	return r
}
