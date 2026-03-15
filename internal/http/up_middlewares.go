package http

import (
	"Alice088/pdf-summarize/pkg/env"
	"Alice088/pdf-summarize/pkg/size"
	"log/slog"

	middlewarex "Alice088/pdf-summarize/internal/http/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	slogchi "github.com/samber/slog-chi"
)

func UpMiddlewares(r *chi.Mux, cfg env.Config, logger *slog.Logger) {
	r.Use(slogchi.New(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Compress(5))
	r.Use(middleware.Throttle(cfg.MaxUser))
	r.Use(middleware.Timeout(cfg.Timeout))
	r.Use(middleware.RequestSize(size.MB5))
	r.Use(middleware.AllowContentEncoding(cfg.AllowContentEncoding...))
	r.Use(middlewarex.PrometheusHttpRequestTotal)
	r.Use(middlewarex.HttpRequestsInFlight)
	r.Use(middlewarex.PrometheusHttpRequestTotalDuration)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: cfg.Origins,
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
