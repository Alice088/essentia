package app

import (
	"Alice088/pdf-summarize/internal/app/dependencies"
	"Alice088/pdf-summarize/internal/controller/restapi"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"Alice088/pdf-summarize/pkg/env"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Run(cfg *env.Config) {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM)

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

	ctxTimeout, cancel := context.WithTimeout(ctx, cfg.DB.OperationTimeout)
	defer cancel()

	conn, err := pgx.Connect(ctxTimeout, cfg.DB.DatabaseURL)
	if err != nil {
		logger.Error("Unable to connect to database", "error", err.Error())
		os.Exit(1)
	}
	defer func() {
		ctx, cancel = context.WithTimeout(context.Background(), cfg.DB.OperationTimeout)
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

	ctxTimeout, cancel = context.WithTimeout(ctx, cfg.MinIO.OperationTimeout)
	defer cancel()

	err = minioClient.MakeBucket(ctxTimeout, "pdf", minio.MakeBucketOptions{Region: cfg.MinIO.Location})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(ctxTimeout, "pdf")
		if errBucketExists != nil || !exists {
			logger.Error("Failed to check bucket exist or bucket doesn't exist", "error", err.Error())
			os.Exit(1)
		}
	}

	r := chi.NewRouter()

	deps := dependencies.AppDeps{
		Config:  cfg,
		Logger:  logger,
		MinIO:   minioClient,
		Queries: q,
	}
	restapi.NewRouter(r, deps)

	srv := &http.Server{
		Addr:         ":" + cfg.HTTP.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	logger.Info("Server starting", "port", cfg.HTTP.Port)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(http.ErrServerClosed, err) {
			logger.Error("server failed", "error", err)
		}
	}()

	<-ctx.Done()
	stop()
	logger.Info("shutting down server...")

	ctxTimeout, cancel = context.WithTimeoutCause(context.Background(), 5*time.Second, errors.New("shutdown too long"))
	defer cancel()

	err = srv.Shutdown(ctxTimeout)
	if err != nil {
		logger.Error("Failed to shutdown server", "error", err.Error())
		os.Exit(1)
	}
}
