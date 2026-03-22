package env

import "time"

type Config struct {
	Env   string `env:"ENV,required"`
	HTTP  HTTP
	DB    DB
	MinIO MinIO
}

type HTTP struct {
	MaxUser              int           `env:"MAX_USERS,required"`
	Timeout              time.Duration `env:"TIMEOUT,required"`
	AllowContentEncoding []string      `env:"ALLOW_CONTENT_ENCODING,required"`
	Origins              []string      `env:"ORIGINS,required"`
	Port                 string        `env:"PORT,required"`
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
	PDFBucket string `env:"MINIO_PDF_BUCKET,required"`
}
