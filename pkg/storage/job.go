package storage

import "github.com/google/uuid"

type JobStage string

const (
	JobStageParse      JobStage = "parse"
	JobStageSplit      JobStage = "split"
	JobStageChunk      JobStage = "chunk"
	JobStageExtract    JobStage = "extract"
	JobStageAggregate  JobStage = "aggregate"
	JobStageGAggregate JobStage = "g_aggregate"
	JobStageBuild      JobStage = "build"
)

type Job struct {
	ID    uuid.UUID
	Stage JobStage
}
