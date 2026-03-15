-------------------------------------------------------------------------------

CREATE TYPE job_status AS ENUM ('uploaded', 'processing', 'completed', 'failed');

CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    status job_status NOT NULL DEFAULT 'uploaded',
    object_key TEXT NOT NULL UNIQUE CHECK (object_key <> ''),
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_status_created ON jobs (status, created_at);

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

-------------------------------------------------------------------------------