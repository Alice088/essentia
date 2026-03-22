package workers

import (
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
