package repo

import (
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"Alice088/essentia/pkg/s3"
	"context"

	"github.com/google/uuid"
)

type Fail struct {
	JobId     uuid.UUID
	Error     error
	ErrorType queries.ErrorType
}

type Job interface {
	CreateJob(ctx context.Context, file s3.File) (uuid.UUID, error)
	FailJob(ctx context.Context, fail Fail) error
	SetJobStage(ctx context.Context, id uuid.UUID, stage string) error
	SetJobText(ctx context.Context, id uuid.UUID, object s3.Object) error
	AdvanceJobStage(ctx context.Context, id uuid.UUID, stage string) error
}
