package env

import "time"

type Config struct {
	Env     string `env:"ENV,required"`
	HTTP    HTTP
	DB      DB
	MinIO   MinIO
	Workers Workers
}

type HTTP struct {
	MaxUser                int           `env:"HTTP_MAX_USERS,required"`
	Timeout                time.Duration `env:"HTTP_TIMEOUT,required"`
	AllowContentEncoding   []string      `env:"HTTP_ALLOW_CONTENT_ENCODING,required"`
	Origins                []string      `env:"HTTP_ORIGINS,required"`
	Port                   string        `env:"HTTP_PORT,required"`
	RateLimitRequestsCount int           `env:"HTTP_RATE_LIMIT_REQUESTS_COUNT,required"`
	RateLimitPerSecond     time.Duration `env:"HTTP_RATE_LIMIT_PER_SECOND,required"`
}

type DB struct {
	DatabaseURL      string        `env:"DATABASE_URL,required"`
	OperationTimeout time.Duration `env:"DATABASE_OPERATION_TIMEOUT,required"`
}

type MinIO struct {
	Endpoint         string        `env:"MINIO_ENDPOINT,required"`
	AccessKey        string        `env:"MINIO_ACCESS_KEY,required"`
	SecretKey        string        `env:"MINIO_SECRET_KEY,required"`
	SSL              bool          `env:"MINIO_SSL,required"`
	Location         string        `env:"MINIO_LOCATION,required"`
	OperationTimeout time.Duration `env:"MINIO_OPERATION_TIMEOUT,required"`
}

type Workers struct {
	WorkerPoolMax int `env:"WORKERS_WORKER_POOL_MAX,required"`
	Parsing       WorkersParsing
}

type WorkersParsing struct {
	ContextTimeout       time.Duration `env:"WORKERS_PARSING_CONTEXT_TIMEOUT,required"`
	ReaderContextTimeout time.Duration `env:"WORKERS_PARSING_READER_CONTEXT_TIMEOUT,required"`
}
