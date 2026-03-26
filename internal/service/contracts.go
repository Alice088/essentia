package service

import (
	"Alice088/essentia/pkg/s3"
	"context"

	"github.com/google/uuid"
)

//go:generate mockery --name=PDF --output=./mocks --outpkg=mocks
type PDF interface {
	Enqueue(ctx context.Context, file s3.File) (uuid.UUID, error)
}
