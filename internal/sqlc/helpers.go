package sqlc

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type JobStage string

const (
	JobStageUploaded JobStage = "uploaded"

	JobStageParsing JobStage = "parsing"
	JobStageParsed  JobStage = "parsed"

	JobStageCleaning JobStage = "cleaning"
	JobStageCleaned  JobStage = "cleaned"

	JobStageChunking JobStage = "chunking"
	JobStageChunked  JobStage = "chunked"

	JobStageSummarizing JobStage = "summarizing"
	JobStageSummarized  JobStage = "summarized"

	JobStageAggregating JobStage = "aggregating"
	JobStageCompleted   JobStage = "completed"
)

type WorkStatus string

const (
	WorkStatusPending    WorkStatus = "pending"
	WorkStatusProcessing WorkStatus = "processing"
	WorkStatusCompleted  WorkStatus = "completed"
	WorkStatusFailed     WorkStatus = "failed"
)

func ToUUID(target uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: target,
		Valid: true,
	}
}

func ToTEXT(target string) pgtype.Text {
	return pgtype.Text{
		String: target,
		Valid:  true,
	}
}
