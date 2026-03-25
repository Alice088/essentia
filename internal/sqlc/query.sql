-- name: CreateJob :one
INSERT INTO jobs (id, object_key)
VALUES ($1, $2) RETURNING *;

-- name: GetJob :one
SELECT *
FROM jobs
WHERE id = $1;

-- name: GetJobByObjectKey :one
SELECT *
FROM jobs
WHERE object_key = $1;

-- name: ListJobs :many
SELECT *
FROM jobs
ORDER BY created_at DESC LIMIT $1
OFFSET $2;

-- name: DeleteJob :exec
DELETE
FROM jobs
WHERE id = $1;


-- name: ClaimNextJobForStage :one
WITH cte AS (
    SELECT id
    FROM jobs
    WHERE jobs.status IN ('pending', 'failed')
      AND jobs.stage = $1
      AND jobs.attempts < 3
      AND (
          jobs.error_type IS NULL
          OR jobs.error_type = ANY($2::text[]::error_type[])
      )
    ORDER BY created_at
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
UPDATE jobs j
SET status = 'processing',
    updated_at = NOW()
FROM cte
WHERE j.id = cte.id
RETURNING
    j.id,
    j.object_key;


-- name: SetJobStage :exec
UPDATE jobs
SET stage = $1
WHERE id = $2;


-- name: AdvanceJobStage :exec
UPDATE jobs
SET stage  = $2,
    status = 'pending',
    error  = NULL,
    error_type = NULL
WHERE id = $1;


-- name: CompleteJob :exec
UPDATE jobs
SET stage  = 'completed',
    status = 'completed',
    error  = NULL,
    error_type = NULL
WHERE id = $1;


-- name: FailJob :exec
UPDATE jobs
SET status   = 'failed',
    error    = $2,
    error_type = $3,
    attempts = attempts + 1
WHERE id = $1;


-- name: SetTextKey :exec
UPDATE jobs
SET text_key = $2
WHERE id = $1;

-- name: SetCleanedTextKey :exec
UPDATE jobs
SET cleaned_text_key = $2
WHERE id = $1;

-- name: SetSummaryKey :exec
UPDATE jobs
SET summary_key = $2
WHERE id = $1;


-- name: CreateChunkTask :one
INSERT INTO chunk_tasks (id,
                         job_id,
                         chunk_index,
                         chunk_key)
VALUES ($1, $2, $3, $4) RETURNING *;


-- name: CreateChunkTasksBatch :exec
INSERT INTO chunk_tasks (id, job_id, chunk_index, chunk_key)
SELECT UNNEST($1::uuid[]),
       $2,
       UNNEST($3::int[]),
       UNNEST($4::text[]);


-- name: GetNextChunkTask :one
SELECT *
FROM chunk_tasks
WHERE status = 'pending'
  AND job_id = $1
ORDER BY chunk_index LIMIT 1
FOR
UPDATE SKIP LOCKED;

-- name: GetNextChunkTaskGlobal :one
SELECT *
FROM chunk_tasks
WHERE status = 'pending'
ORDER BY created_at
LIMIT 1
FOR UPDATE SKIP LOCKED;

-- name: SetChunkProcessing :exec
UPDATE chunk_tasks
SET status = 'processing'
WHERE id = $1;

-- name: CompleteChunkTask :exec
UPDATE chunk_tasks
SET status = 'completed',
    result_key = $2,
    error = NULL
WHERE id = $1;


-- name: FailChunkTask :exec
UPDATE chunk_tasks
SET status = 'failed',
    error = $2,
    attempts = attempts + 1
WHERE id = $1;

-- name: CountPendingChunks :one
SELECT COUNT(*)
FROM chunk_tasks
WHERE job_id = $1
  AND status != 'completed';


-- name: ResetFailedChunks :exec
UPDATE chunk_tasks
SET status = 'pending',
    error = NULL
WHERE job_id = $1
  AND status = 'failed';

-- name: ResetFailedJob :exec
UPDATE jobs
SET status = 'pending',
    error = NULL,
    error_type = NULL
WHERE id = $1
  AND status = 'failed';
