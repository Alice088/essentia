package main

import (
	"Alice088/essentia/pkg/env"
	"context"
	"log"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	cfg := env.Load("./.env")

	client, err := minio.New(cfg.MinIO.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIO.AccessKey, cfg.MinIO.SecretKey, ""),
		Secure: cfg.MinIO.SSL,
	})
	if err != nil {
		panic(err)
	}

	bucket := "pdf"
	ctx := context.Background()

	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Println("list error:", object.Err)
				continue
			}
			objectsCh <- object
		}
	}()

	for rErr := range client.RemoveObjects(ctx, bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		log.Println("remove error:", rErr)
	}

	log.Println("bucket cleaned")
}
