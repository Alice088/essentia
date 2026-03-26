package job

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

func NewRepo(deps dependencies.AppDeps) repo.Job {
	return basic{
		Queries: deps.Queries,
		Config:  deps.Config.DB,
	}
}

func (b basic) AdvanceJobStage(ctx context.Context, id uuid.UUID, stage string) error {
	timeout, cancel := context.WithTimeout(ctx, b.Config.OperationTimeout)
	defer cancel()
	err := b.Queries.AdvanceJobStage(timeout, queries.AdvanceJobStageParams{
		ID:    sqlc.ToUUID(id),
		Stage: queries.JobStage(stage),
	})
	if err != nil {
		return fmt.Errorf("failed to advance job: %w", err)
	}
	return nil
}

func (b basic) SetJobText(ctx context.Context, id uuid.UUID, text string) error {
	timeout, cancel := context.WithTimeout(ctx, b.Config.OperationTimeout)
	defer cancel()
	err := b.Queries.SetTextKey(timeout, queries.SetTextKeyParams{
		ID:      sqlc.ToUUID(id),
		TextKey: sqlc.ToTEXT(text),
	})
	if err != nil {
		return fmt.Errorf("failed to set text key: %w", err)
	}
	return nil
}

func (b basic) SetJobStage(ctx context.Context, id uuid.UUID, stage string) error {
	timeout, cancel := context.WithTimeout(ctx, b.Config.OperationTimeout)
	defer cancel()
	err := b.Queries.SetJobStage(timeout, queries.SetJobStageParams{
		ID:    sqlc.ToUUID(id),
		Stage: queries.JobStage(stage),
	})
	if err != nil {
		return fmt.Errorf("failed to set job state: %w", err)
	}
	return nil
}

func (b basic) FailJob(ctx context.Context, fail repo.Fail) error {
	timeout, cancel := context.WithTimeout(ctx, b.Config.OperationTimeout)
	defer cancel()

	err := b.Queries.FailJob(timeout, queries.FailJobParams{
		ID:    sqlc.ToUUID(fail.JobId),
		Error: sqlc.ToTEXT(fail.Error.Error()),
		ErrorType: queries.NullErrorType{
			ErrorType: fail.ErrorType,
			Valid:     true,
		},
	})

	if err != nil {
		return fmt.Errorf("fail to fail job: %w", err)
	}
	return nil
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
