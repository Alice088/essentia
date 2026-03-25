-------------------------------------------------------------------------------
CREATE TYPE job_stage AS ENUM (
    'uploaded',

    'parsing',
    'parsed',

    'cleaning',
    'cleaned',

    'chunking',
    'chunked',

    'summarizing',
    'summarized',

    'aggregating',
    'completed'
);

CREATE TYPE work_status AS ENUM ('pending', 'processing', 'completed', 'failed');

CREATE TYPE parsing_error_type AS ENUM (
    'open',
    'corrupted',
    'encrypted',
    'timeout',
    'extract',
    'storage_download',
    'storage_upload',
    'db',
    'unknown'
);

CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    stage job_stage NOT NULL DEFAULT 'uploaded',
    status work_status NOT NULL DEFAULT 'pending',
    object_key TEXT NOT NULL UNIQUE CHECK (object_key <> ''),
    attempts INT NOT NULL DEFAULT 0,
    text_key TEXT,
    cleaned_text_key TEXT,
    summary_key TEXT,
    error TEXT,
    error_type parsing_error_type,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_status_stage_created
ON jobs (status, stage, created_at);

-- автоматическое обновление updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_jobs_updated_at
BEFORE UPDATE ON jobs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


CREATE TABLE chunk_tasks (
    id UUID PRIMARY KEY,
    job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL,
    status work_status NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    -- ссылка на chunk (в MinIO)
    chunk_key TEXT NOT NULL,
    -- ссылка на результат
    result_key TEXT,
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (job_id, chunk_index)
);

CREATE INDEX idx_chunk_tasks_job_status
ON chunk_tasks (job_id, status);

CREATE INDEX idx_chunk_tasks_status
ON chunk_tasks (status, created_at);

CREATE TRIGGER trg_chunks_tasks_updated_at
BEFORE UPDATE ON chunk_tasks
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-------------------------------------------------------------------------------
