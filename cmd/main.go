package main

import (
	httpx "Alice088/pdf-summarize/internal/http"
	v1 "Alice088/pdf-summarize/internal/http/v1"
	"Alice088/pdf-summarize/pkg/env"
	"Alice088/pdf-summarize/prometheus"

	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	cfg := env.Load("./.env")

	logRotator := &lumberjack.Logger{
		Filename:   "./logs/logs.log",
		MaxSize:    10,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}

	mw := slog.NewJSONHandler(
		io.MultiWriter(os.Stdout, logRotator),
		nil,
	)
	logger := slog.New(mw)

	r := chi.NewRouter()
	httpx.UpMiddlewares(r, cfg, logger)
	prometheus.UpMetrics()

	r.Mount("/v1", v1.Routes(logger))
	r.Handle("/metrics", promhttp.Handler())

	http.ListenAndServe(":3000", r)
}
