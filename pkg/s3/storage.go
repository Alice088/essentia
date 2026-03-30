package s3

import (
	"context"
)

type S3 interface {
	FilesManager
}

type FilesManager interface {
	Put(ctx context.Context, file File) error
	Get(ctx context.Context, file File) error
	Delete(ctx context.Context, file File) error
}
