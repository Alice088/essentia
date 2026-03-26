package repo

import (
	"Alice088/essentia/pkg/s3"
	"context"

	"github.com/google/uuid"
)

type PDF interface {
	CreateJob(ctx context.Context, file s3.File) (uuid.UUID, error)
}
