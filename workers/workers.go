package workers

import (
	"Alice088/pdf-summarize/internal/dependencies"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"Alice088/pdf-summarize/pkg/env"
	"log/slog"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type Task struct {
	UUID      uuid.UUID
	ObjectKey string
}

type Worker struct {
	Logger  *slog.Logger
	Queries *queries.Queries
	MinIO   *minio.Client
	Config  env.Workers
}

func NewWorker(deps dependencies.AppDeps) *Worker {
	return &Worker{
		Logger:  deps.Logger,
		Queries: deps.Queries,
		MinIO:   deps.MinIO,
		Config:  deps.Config.Workers,
	}
}
