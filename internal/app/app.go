package app

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/controller/restapi"
	"Alice088/essentia/internal/repo/job"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"Alice088/essentia/internal/workers"
	"Alice088/essentia/pkg/env"
	"Alice088/essentia/pkg/pdf_parser"
	"Alice088/essentia/pkg/wminio"
	"Alice088/essentia/pkg/xlogger"
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Run(cfg env.Config) {
	ctx, stop := signal.NotifyContext(
		context.Background(), syscall.SIGINT, syscall.SIGTERM,
	)

	logger := xlogger.New(cfg)

	ctxTimeout, cancel := context.WithTimeout(ctx, cfg.DB.OperationTimeout)
	defer cancel()

	conn, err := pgxpool.New(ctxTimeout, cfg.DB.DatabaseURL)
	if err != nil {
		logger.Error("Unable to connect to database", "error", err.Error())
		return
	}
	defer conn.Close()

	s3, err := wminio.New(cfg.MinIO, "pdf")
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		return
	}

	err = s3.CreateBucketIfNotExists(ctx)
	if err != nil {
		logger.Error(err.Error(), "error", errors.Unwrap(err))
		return
	}

	deps := dependencies.AppDeps{
		Config:  cfg,
		Logger:  logger,
		S3:      s3,
		Queries: queries.New(conn),
		DB:      conn,
	}
	deps.JobRepo = job.NewRepo(deps)

	r := chi.NewRouter()
	restapi.NewRouter(r, deps)

	parser := workers.Parser{
		Repo:      deps.JobRepo,
		PDFParser: pdf_parser.NewParser(cfg.Workers.Parsing),
		Logger:    logger,
		S3:        s3,
	}

	wgConsumer := &sync.WaitGroup{}
	wgProducer := &sync.WaitGroup{}

	tasks := make(chan workers.Job, 4)
	workers.UpConsumerWorkerPool(deps, wgConsumer, workers.ConsumerWorkerPoolConfig{
		WorkerName:   "ParsingWorker",
		Timeout:      cfg.Workers.Parsing.ContextTimeout,
		WorkersCount: 2,
		Workers:      parser.Parsing,
		Jobs:         tasks,
	})

	workers.UpStreamWorkerPool(deps, wgProducer, workers.UpStreamWorkerPoolConfig{
		WorkerName:   "StreamParsingJobsWorker",
		Timeout:      cfg.Workers.Parsing.ContextTimeout,
		WorkersCount: 2,
		Worker:       workers.StreamParsingJobs,
		Jobs:         tasks,
		GlobalCtx:    ctx,
	})

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

	logger.Info("waiting for stream workers...")
	wgProducer.Wait()
	close(tasks)

	logger.Info("waiting for consumer workers...")
	wgConsumer.Wait()

	logger.Info("finish!")
}
