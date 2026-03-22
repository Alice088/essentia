package service

import (
	"Alice088/pdf-summarize/internal/dependencies"
	queries "Alice088/pdf-summarize/internal/sqlc/postgresql"
	"context"
	"io"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/minio/minio-go/v7"
)

//go:generate mockery --name=PDFService --output=./mocks --outpkg=mocks
type PDFService interface {
	CreateJob(ctx context.Context, r io.Reader, size int64) (uuid.UUID, error)
}

type pdfService struct {
	MinIO   *minio.Client
	Queries *queries.Queries
	Logger  *slog.Logger
}

func NewService(appDeps dependencies.AppDeps) PDFService {
	return &pdfService{
		MinIO:   appDeps.MinIO,
		Queries: appDeps.Queries,
		Logger:  appDeps.Logger,
	}
}

func (s *pdfService) CreateJob(
	ctx context.Context,
	reader io.Reader,
	size int64,
) (uuid.UUID, error) {
	jobID := uuid.New()
	objectKey := jobID.String() + ".pdf"

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
