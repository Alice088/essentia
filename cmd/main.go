package main

import (
	httpx "Alice088/pdf-summarize/internal/http"
	v1 "Alice088/pdf-summarize/internal/http/v1"
	"Alice088/pdf-summarize/internal/prometheus"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"Alice088/pdf-summarize/pkg/env"
	"context"
	"errors"
	"os/signal"
	"syscall"
	"time"

	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM)

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

	ctxTimeout, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctxTimeout, cfg.DB.DatabaseURL)
	if err != nil {
		logger.Error("Unable to connect to database", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := conn.Close(ctx); err != nil {
			logger.Error("failed to close db", "error", err)
		}
	}()

	q := queries.New(conn)

	minioClient, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.SSL,
	})
	if err != nil {
		logger.Error("Unable to connect to minio", "error", err.Error())
		os.Exit(1)
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = minioClient.MakeBucket(ctxTimeout, cfg.MinIO.PDFBucket, minio.MakeBucketOptions{Region: cfg.MinIO.Location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctxTimeout, cfg.MinIO.PDFBucket)
		if errBucketExists != nil || !exists {
			logger.Error("Failed to check bucket exist or bucket doesn't exist", "error", err.Error())
			os.Exit(1)
		}
	}

	r := chi.NewRouter()
	httpx.UpMiddlewares(r, cfg, logger)
	prometheus.UpMetrics()

	r.Mount("/v1", v1.Routes(logger, q, cfg.HTTP.Timeout, minioClient, cfg.MinIO.PDFBucket))
	r.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logger.Error("server failed", "error", err)
		}
	}()

	<-ctx.Done()
	stop()
	logger.Info("shutting down server...")

	ctxTimeout, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = srv.Shutdown(ctxTimeout)
	if err != nil {
		logger.Error("Failed to shutdown server", "error", err.Error())
		os.Exit(1)
	}
}
