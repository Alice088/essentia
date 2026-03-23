package pdf

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/service"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"
)

type basic struct {
	Deps *dependencies.AppDeps
}

func New(appDeps *dependencies.AppDeps) service.PDF {
	return &basic{
		Deps: appDeps,
	}
}

func (s *basic) Enqueue(
	ctx context.Context,
	reader io.Reader,
	size int64,
) (uuid.UUID, error) {
	jobID := uuid.New()
	objectKey := jobID.String() + ".pdf"

	uploaded := false
	defer func() {
		if !uploaded {
			ctxTimeout, cancel := context.WithTimeout(context.Background(), s.Deps.Config.MinIO.OperationTimeout)
			defer cancel()

			err := s.Deps.MinIO.RemoveObject(ctxTimeout, "pdf", objectKey, minio.RemoveObjectOptions{})
			if err != nil {
				s.Deps.Logger.Error("failed to remove object", "object", objectKey, "error", err)
			}
		}
	}()

	ctxTimeout, cancel := context.WithTimeout(ctx, s.Deps.Config.MinIO.OperationTimeout)
	defer cancel()
	_, err := s.Deps.MinIO.PutObject(
		ctxTimeout,
		"pdf",
		objectKey,
		reader,
		size,
		minio.PutObjectOptions{
			ContentType: "application/pdf",
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	ctxTimeout, cancel = context.WithTimeout(ctx, s.Deps.Config.DB.OperationTimeout)
	defer cancel()
	_, err = s.Deps.Queries.CreateJob(ctxTimeout, queries.CreateJobParams{
		ID: pgtype.UUID{
			Bytes: jobID,
			Valid: true,
		},
		ObjectKey: objectKey,
	})
	if err != nil {
		return uuid.Nil, err
	}

	uploaded = true
	return jobID, nil
}
