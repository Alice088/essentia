package pdf

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/repo"
	"Alice088/essentia/internal/service"
	"Alice088/essentia/pkg/s3"
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
)

type basic struct {
	S3     s3.S3
	Logger *slog.Logger
	Repo   repo.PDF
}

func New(deps dependencies.AppDeps, repo repo.PDF) service.PDF {
	return &basic{
		S3:     deps.S3,
		Logger: deps.Logger,
		Repo:   repo,
	}
}

func (s *basic) Enqueue(
	ctx context.Context,
	file s3.File,
) (uuid.UUID, error) {
	file.Object = s3.NewPDF()

	uploaded := false
	defer func() {
		if !uploaded {
			err := s.S3.Delete(ctx, file.Object)
			if err != nil {
				s.Logger.Error(err.Error(), "error", errors.Unwrap(err))
			}
		}
	}()

	err := s.S3.Put(ctx, file)
	if err != nil {
		return uuid.Nil, err
	}

	id, err := s.Repo.CreateJob(ctx, file)
	if err != nil {
		return uuid.Nil, err
	}

	uploaded = true
	return id, nil
}
