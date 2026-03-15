-- name: CreateJob :one
INSERT INTO jobs (id, object_key) VALUES ($1, $2) RETURNING *;

-- name: GetJob :one
SELECT * FROM jobs WHERE id = $1;

-- name: GetJobByObjectKey :one
SELECT * FROM jobs WHERE object_key = $1;

-- name: ListJobs :many
SELECT * FROM jobs ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: GetNextUploadedJob :one
SELECT *
FROM jobs
WHERE
    status = 'uploaded'
ORDER BY created_at
LIMIT 1 FOR
UPDATE SKIP LOCKED;

-- name: SetJobProcessing :exec
UPDATE jobs SET status = 'processing' WHERE id = $1;

-- name: SetJobCompleted :exec
UPDATE jobs SET status = 'completed', error = NULL WHERE id = $1;

-- name: SetJobFailed :exec
UPDATE jobs SET status = 'failed', error = $2 WHERE id = $1;

-- name: DeleteJob :exec
DELETE FROM jobs WHERE id = $1;