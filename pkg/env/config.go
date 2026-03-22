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
	MaxUser              int           `env:"HTTP_MAX_USERS,required"`
	Timeout              time.Duration `env:"HTTP_TIMEOUT,required"`
	AllowContentEncoding []string      `env:"HTTP_ALLOW_CONTENT_ENCODING,required"`
	Origins              []string      `env:"HTTP_ORIGINS,required"`
	Port                 string        `env:"HTTP_PORT,required"`
}

type DB struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
}

type MinIO struct {
	Endpoint  string `env:"MINIO_ENDPOINT,required"`
	AccessKey string `env:"MINIO_ACCESS_KEY,required"`
	SecretKey string `env:"MINIO_SECRET_KEY,required"`
	SSL       bool   `env:"MINIO_SSL,required"`
	Location  string `env:"MINIO_LOCATION,required"`
}

type Workers struct {
	ContextTimeout time.Duration `env:"WORKERS_CONTEXT_TIMEOUT,required"`
	WorkerPoolMax  int           `env:"WORKERS_WORKER_POOL_MAX,required"`
}
