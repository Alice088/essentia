package storage

import (
	"context"

	"github.com/google/uuid"
)

type Storage interface {
	GetProcessableJobs(ctx context.Context, limit int) []Job
	InvalidJob(context.Context, uuid.UUID) error
}
