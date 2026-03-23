package service

import (
	"context"
	"io"

	"github.com/google/uuid"
)

//go:generate mockery --name=PDF --output=./mocks --outpkg=mocks
type PDF interface {
	Enqueue(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error)
}
