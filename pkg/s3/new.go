package s3

import (
	"Alice088/essentia/pkg/pdf_parser"
	"context"
	"io"
)

type S3 interface {
	BuckerManager
	FilesManager
}

type BuckerManager interface {
	BucketExists(ctx context.Context) (bool, error)
	CreateBucket(ctx context.Context) error
	CreateBucketIfNotExists(ctx context.Context) error
}

type File struct {
	Object Object
	Reader io.Reader
	Size   int64
}

type FilesManager interface {
	Put(ctx context.Context, file File) error
	FGet(ctx context.Context, object Object, tmp pdf_parser.TMP) error
	Delete(ctx context.Context, object Object) error
}
