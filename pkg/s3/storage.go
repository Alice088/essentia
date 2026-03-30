package storage

import (
	"context"
)

type Storage interface {
	FilesManager
}

type FilesManager interface {
	Put(ctx context.Context, file File) error
	Get(ctx context.Context, file File) error
	Delete(ctx context.Context, file File) error
}
