package restapi

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/controller/restapi/middleware"
	"Alice088/essentia/internal/controller/restapi/v1"
	"Alice088/essentia/pkg/prometheus"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewRouter(r *chi.Mux, deps *dependencies.AppDeps) {
	middleware.UpMiddlewares(r, deps.Config, deps.Logger)
	prometheus.UpMetrics()

	r.Handle("/metrics", promhttp.Handler())
	r.Mount("/v1", v1.Routes(deps))
}
