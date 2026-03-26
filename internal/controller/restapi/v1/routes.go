package v1

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/controller/restapi/v1/pdf"
	pdfservice "Alice088/essentia/internal/service/pdf"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Routes(deps dependencies.AppDeps) chi.Router {
	r := chi.NewRouter()

	pdfHandler := pdf.NewHandler(deps, pdfservice.New(deps))
	r.With(middleware.AllowContentType("application/pdf")).Post("/pdf/load", pdfHandler.Load())

	return r
}
