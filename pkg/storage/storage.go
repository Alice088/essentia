package storage

import (
	"context"

	"github.com/google/uuid"
)

type Storage interface {
	GetProcessableJobs(context.Context) []Job
	InvalidJob(context.Context, uuid.UUID) error
}
