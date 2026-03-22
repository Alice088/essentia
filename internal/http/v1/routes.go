package v1

import (
	"Alice088/pdf-summarize/internal/dependencies"
	"Alice088/pdf-summarize/internal/http/v1/pdf"
	"Alice088/pdf-summarize/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Routes(appDeps dependencies.AppDeps) chi.Router {
	r := chi.NewRouter()

	pdfHandler := pdf.NewHandler(appDeps, service.NewPDFService(appDeps))

	r.With(middleware.AllowContentType("application/pdf")).Post("/load", pdfHandler.Load())
	return r
}
