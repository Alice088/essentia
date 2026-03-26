package wminio

import (
	"Alice088/essentia/pkg/env"
	"Alice088/essentia/pkg/s3"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Wrapper struct {
	client *minio.Client
	bucket string
	config env.MinIO
}

func (w Wrapper) Delete(ctx context.Context, object s3.Object) error {
	timeout, cancel := context.WithTimeout(ctx, w.config.OperationTimeout)
	defer cancel()
	err := w.client.RemoveObject(timeout, w.bucket, object.Key(), minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete object failed: %w", err)
	}
	return nil
}

func (w Wrapper) CreateBucketIfNotExists(ctx context.Context) error {
	exist, err := w.BucketExists(ctx)
	if err != nil {
		return err
	}

	if !exist {
		return w.CreateBucket(ctx)
	}

	return nil
}

func (w Wrapper) BucketExists(ctx context.Context) (bool, error) {
	timeout, cancel := context.WithTimeout(ctx, w.config.OperationTimeout)
	defer cancel()

	exists, err := w.client.BucketExists(timeout, w.bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket exist: %w", err)
	}

	return exists, nil
}

func (w Wrapper) CreateBucket(ctx context.Context) error {
	timeout, cancel := context.WithTimeout(ctx, w.config.OperationTimeout)
	defer cancel()

	err := w.client.MakeBucket(timeout, w.bucket, minio.MakeBucketOptions{Region: w.config.Location})
	if err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

func (w Wrapper) Put(ctx context.Context, file s3.File) error {
	timeout, cancel := context.WithTimeout(ctx, w.config.OperationTimeout)
	defer cancel()

	_, err := w.client.PutObject(
		timeout,
		w.bucket,
		file.Object.Key(),
		file.Reader,
		file.Size,
		minio.PutObjectOptions{
			ContentType: "application/pdf",
		},
	)
	if err != nil {
		return fmt.Errorf("failed put object: %w", err)
	}
	return nil
}

func (w Wrapper) Get(ctx context.Context, object s3.Object, tmp string) error {
	//TODO implement me
	panic("implement me")
}

func New(cfg env.MinIO, bucket string) (s3.S3, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		return Wrapper{}, fmt.Errorf("unable to connect to minio: %w", err)
	}

	return Wrapper{client: client, bucket: bucket, config: cfg}, nil
}
