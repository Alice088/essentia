package middleware

import (
	"Alice088/pdf-summarize/pkg/env"
	"Alice088/pdf-summarize/pkg/size"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	slogchi "github.com/samber/slog-chi"
)

func UpMiddlewares(r *chi.Mux, cfg *env.Config, logger *slog.Logger) {
	r.Use(slogchi.NewWithFilters(
		logger,
		slogchi.IgnorePath("/metrics"),
	))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Throttle(cfg.HTTP.MaxUser))
	r.Use(middleware.Timeout(cfg.HTTP.Timeout))
	r.Use(middleware.RequestSize(size.MB5))
	r.Use(middleware.AllowContentEncoding(cfg.HTTP.AllowContentEncoding...))
	r.Use(PrometheusHttpRequestTotal)
	r.Use(HttpRequestsInFlight)
	r.Use(PrometheusHttpRequestTotalDuration)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.HTTP.Origins,
		AllowedMethods: []string{
			"GET",
			"POST",
			"OPTIONS",
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
		},
		ExposedHeaders: []string{
			"Content-Length",
		},
		AllowCredentials: false,
		MaxAge:           300,
	}))

}
