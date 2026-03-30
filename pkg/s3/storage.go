package s3

import (
	"Alice088/essentia/internal/domain/pipeline"
	"context"
)

type S3 interface {
	FilesManager
}

type FilesManager interface {
	Put(ctx context.Context, file File) error
	Get(ctx context.Context, file File) error
	Find(ctx context.Context, name, bucket string) ([]pipeline.Blob, error)
	Delete(ctx context.Context, file File) error
}
