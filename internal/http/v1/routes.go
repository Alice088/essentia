package v1

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Routes() chi.Router {
	r := chi.NewRouter()

	r.With(middleware.AllowContentType("application/pdf")).Post("/load", Load())
	return r
}
