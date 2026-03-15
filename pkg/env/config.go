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
}

type DB struct {
	DatabaseURL string `env:"DATABASE_URL,required"`
}

type MinIO struct {
	Endpoint  string `env:"ENDPOINT,required"`
	AccessKey string `env:"ACCESS_KEY,required"`
	SecretKey string `env:"SECRET_KEY,required"`
	SSL       bool   `env:"SSL,required"`
	Location  string `env:"LOCATION,required"`
	PDFBucket string `env:"PDF_BUCKET,required"`
}
