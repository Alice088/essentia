package pdf

import (
	"Alice088/essentia/internal/app/dependencies"
	"Alice088/essentia/internal/repo"
	"Alice088/essentia/internal/sqlc"
	queries "Alice088/essentia/internal/sqlc/postgresql"
	"Alice088/essentia/pkg/env"
	"Alice088/essentia/pkg/s3"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type basic struct {
	Queries *queries.Queries
	Config  env.DB
}

func NewRepo(deps dependencies.AppDeps) repo.PDF {
	return basic{
		Queries: deps.Queries,
		Config:  deps.Config.DB,
	}
}

func (b basic) CreateJob(ctx context.Context, file s3.File) (uuid.UUID, error) {
	timeout, cancel := context.WithTimeout(ctx, b.Config.OperationTimeout)
	defer cancel()

	id, err := b.Queries.CreateJob(timeout, queries.CreateJobParams{
		ID:        sqlc.ToUUID(file.Object.Name),
		ObjectKey: file.Object.Key(),
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed create job: %w", err)
	}

	return uuid.MustParse(id.String()), nil
}
