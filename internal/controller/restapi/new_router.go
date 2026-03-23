package restapi

import (
	"Alice088/pdf-summarize/internal/app/dependencies"
	"Alice088/pdf-summarize/internal/controller/restapi/middleware"
	"Alice088/pdf-summarize/internal/controller/restapi/v1"
	"Alice088/pdf-summarize/pkg/prometheus"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(r *chi.Mux, deps *dependencies.AppDeps) {
	middleware.UpMiddlewares(r, deps.Config, deps.Logger)
	prometheus.UpMetrics()

	r.Handle("/metrics", promhttp.Handler())
	r.Mount("/v1", v1.Routes(deps))
}
