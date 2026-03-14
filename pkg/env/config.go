package env

import "time"

type Config struct {
	Env                  string        `env:"ENV,required"`
	MaxUser              int           `env:"MAX_USERS,required"`
	Timeout              time.Duration `env:"TIMEOUT,required"`
	AllowContentEncoding []string      `env:"ALLOW_CONTENT_ENCODING,required"`
	Origins              []string      `env:"ORIGINS,required"`
}
