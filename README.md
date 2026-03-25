# Essentia — AI PDF Summarization Service

Essentia is a backend service that accepts a PDF file, extracts text, and prepares data for AI summarization.

Goal: give the user a short, useful summary of document content.

## What the service does

- Accepts PDF via HTTP API
- Validates input (format + size)
- Stores source file in object storage
- Runs asynchronous processing pipeline (workers)
- Exposes Prometheus metrics for load and performance analysis

## Processing pipeline

Current target pipeline:

1. PDF upload ✅
2. Basic validation ✅
3. Store file ✅
4. PDF parsing ✅
5. Text cleaning 🚧
6. Chunking 🚧
7. LLM summarization 🚧
8. Summary aggregation 🚧

`✅` — implemented, `🚧` — planned/in progress.

## Core functionality

### 1) Document intake

- `POST /v1/pdf/load`
- Required content type: `application/pdf`
- Validates:
  - file is PDF
  - max upload size is 5 MB

### 2) Document processing

- Extracts text from PDF (sandboxed reader process)
- Handles parsing failures with typed error codes
- Stores parsing results and job state in DB

### 3) AI processing (target)

- Split text into chunks
- Send chunks to LLM
- Aggregate partial summaries into final summary

### 4) Result delivery

- API returns `job_id` for uploaded document
- Final summary endpoint is planned as next stage

## Constraints

- File size: up to 5 MB
- File format: PDF only
- Service-side processing target: up to 5s (without LLM)
- Concurrent users target: up to 10

## Non-functional requirements

### Performance

- Target latency: `< 200ms` for service-level HTTP handling (without LLM)

### Reliability

- Parsing error handling
- Corrupted file protection
- Typed parsing error diagnostics

### Security

- MIME/content-type checks
- Upload size limits
- Isolated PDF reader execution with resource limits

## API

### Upload PDF

`POST /v1/pdf/load`

Example:

```bash
curl -X POST \
  -H "Content-Type: application/pdf" \
  --data-binary "@./sample.pdf" \
  http://localhost:3000/v1/pdf/load
```

Response:

```json
{
  "job_id": "f4c57277-5cb6-4e09-bd76-1b14c2f714a0"
}
```

## Metrics

Metrics endpoint: `GET /metrics`

### Implemented now

- `http_requests_total` — base request counter (`requests/min`, `error rate` calculations)
- `http_requests_in_flight` — active requests (`active users` proxy)
- `http_request_total_duration_seconds` — total request time (`total request time`)
- `parsing_total{status="success|failed"}` — parsing load and success/fail split
- `parsing_duration_seconds{status=...}` — parsing duration for p50/p95/p99
- `parsing_pdf_size_bytes{status=...}` — PDF size distribution and latency correlation
- `parsing_errors_total{error_type=...}` — parsing errors by type
- Go runtime/process metrics (from Prometheus Go client):
  - `go_*` (`go version` / runtime insights)
  - `process_cpu_seconds_total` (`cpu usage` via rate)
  - `process_resident_memory_bytes` (`memory usage`)

Current parsing error types:

- `open`
- `corrupted`
- `encrypted`
- `timeout`
- `extract`
- `empty`
- `storage_download`
- `storage_upload`
- `db`
- `unknown`

### Planned metrics (will expand)

- `service latency` (stage-level split without LLM)
- `llm latency`
- `documents processed`
- `avg pdf pages`
- `disk usage`
- `disk write/read rate`
- `queue size`
- `active jobs`
- `stage latency` per pipeline step
- `llm tokens used`
- `llm requests`
- `worker uptime`

## Example PromQL

Parsing p95:

```promql
histogram_quantile(0.95, sum(rate(parsing_duration_seconds_bucket{status="success"}[5m])) by (le))
```

Parsing success rate:

```promql
sum(rate(parsing_total{status="success"}[5m]))
/
sum(rate(parsing_total[5m]))
```

Top parsing errors:

```promql
topk(5, sum(rate(parsing_errors_total[5m])) by (error_type))
```

Requests per minute:

```promql
sum(increase(http_requests_total[1m]))
```

## Local run

### Prerequisites

- Go `1.26+`
- PostgreSQL
- MinIO
- Linux with `systemd-run` (used for isolated PDF reader process)

### 1) Configure environment

Service reads `./.env`. Required keys include:

- `ENV`
- `HTTP_MAX_USERS`
- `HTTP_TIMEOUT`
- `HTTP_ALLOW_CONTENT_ENCODING`
- `HTTP_ORIGINS`
- `HTTP_PORT`
- `HTTP_RATE_LIMIT_REQUESTS_COUNT`
- `HTTP_RATE_LIMIT_PER_SECOND`
- `DATABASE_URL`
- `DATABASE_OPERATION_TIMEOUT`
- `MINIO_ENDPOINT`
- `MINIO_ACCESS_KEY`
- `MINIO_SECRET_KEY`
- `MINIO_SSL`
- `MINIO_LOCATION`
- `MINIO_OPERATION_TIMEOUT`
- `WORKERS_WORKER_POOL_MAX`
- `WORKERS_PARSING_CONTEXT_TIMEOUT`
- `WORKERS_PARSING_READER_CONTEXT_TIMEOUT`
- `WORKERS_WRITE_TASKS_CONTEXT_TIMEOUT`
- `WORKERS_WRITE_TASKS_ERROR_SLEEP`

### 2) Build PDF reader binary

```bash
./scripts/build_pdf_reader.sh ./cmd/pdf_reader ./build/pdf_reader
```

### 3) Run service

```bash
go run ./cmd/essentia
```

### 4) Run Prometheus (optional)

```bash
docker compose up -d prometheus
```

## Tech stack

### Backend
- **Go** – primary language
- **Chi** – HTTP router for API routes and middleware
- **Chi CORS** – Cross‑Origin Resource Sharing handling
- **github.com/go-chi/httprate** – request rate‑limiting middleware
- **github.com/ledongthuc/pdf** – PDF text extraction library
- **google/uuid** – RFC 9562/DCE 1.1 UUID generation and inspection
- **slog‑chi** – Chi middleware for structured logging (slog)
- **lumberjack** – log rotation for Go

### Database & Storage
- **PostgreSQL** – open‑source relational database with extensibility and SQL compliance
- **sqlc** – type‑safe Go code generation from SQL
- **MinIO** – open‑source, high‑performance S3‑compatible object storage
### Configuration & Environment
- **godotenv** – loads variables from `.env` files into the environment
- **caarlos0/env** – zero‑dependencies environment‑variable parsing into structs

### Observability
- **Prometheus** – open‑source monitoring and alerting toolkit

### Testing & Development
- **vektra/mockery** – mock code autogenerator for Go

### AI & Automation
- **Codex** / **DeepSeek‑chat** – routine coding, code review, and commit‑message generation
- **DeepSeek‑chat** – penetration‑testing assistance and code analysis
- **PentAGI** – fully autonomous AI agents for complex penetration‑testing tasks (https://github.com/vxcontrol/pentagi)
