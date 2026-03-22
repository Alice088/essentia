package service

import (
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"context"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"
)

type PDFService struct {
	MinIO   *minio.Client
	Queries *queries.Queries
	Timeout time.Duration
	Logger  *slog.Logger
}

func (s *PDFService) CreateJobFromPDF(
	ctx context.Context,
	reader io.Reader,
	size int64,
) (uuid.UUID, error) {
	jobID := uuid.New()
	objectKey := jobID.String() + ".pdf"

	ctx, cancel := context.WithTimeout(ctx, s.Timeout)
	defer cancel()

	uploaded := false
	defer func() {
		if !uploaded {
			err := s.MinIO.RemoveObject(ctx, "pdf", objectKey, minio.RemoveObjectOptions{})
			if err != nil {
				s.Logger.Error("failed to remove object", "object", objectKey, "error", err)
			}
		}
	}()

	_, err := s.MinIO.PutObject(
		ctx,
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
	uploaded = true

	_, err = s.Queries.CreateJob(ctx, queries.CreateJobParams{
		ID: pgtype.UUID{
			Bytes: jobID,
			Valid: true,
		},
		ObjectKey: objectKey,
	})
	if err != nil {
		_ = s.MinIO.RemoveObject(ctx, "pdf", objectKey, minio.RemoveObjectOptions{})
		return uuid.Nil, err
	}

	return jobID, nil
}
